package sample

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

var tmpDir string

// https://www.voitanos.io/blog/updated-creating-and-trusting-self-signed-certs-on-macos-and-chrome/
func Run() {
	MakrDirectory()
	// defer os.RemoveAll(tmpDir)
	ChangeDirectory()

	one := `#!/bin/bash
cat << EOF > rootca.conf
[ req ]
prompt             = no
distinguished_name = dn-param
x509_extensions    = ca_cert_extensions

[ ca ]
default_ca = ca_default

[ dn-param ]
C  = US
CN = Root CA

[ ca_cert_extensions ]
keyUsage         = keyCertSign, digitalSignature
basicConstraints = CA:TRUE, pathlen:2

[ ca_default ]
new_certs_dir = .              # Location for new certs after signing
database      = ./index.txt    # Database index file
serial        = ./serial.txt   # The current serial number

default_days  = 1000
default_md    = sha256

policy        = signing_policy
email_in_dn   = no

[ intermediate_cert_extensions ]
keyUsage         = keyCertSign, digitalSignature
basicConstraints = CA:TRUE, pathlen:1

[client_cert_extensions]
keyUsage         = keyCertSign, digitalSignature
basicConstraints = CA:FALSE

[ signing_policy ]
countryName            = optional
stateOrProvinceName    = optional
localityName           = optional
organizationName       = optional
organizationalUnitName = optional
commonName             = supplied
emailAddress           = optional
EOF
`

	two := `#!/bin/bash
cat <<EOF > serverca.conf
[ req ]
default_bits        = 2048
default_keyfile     = rootca.key
distinguished_name  = subject
req_extensions      = req_ext
x509_extensions     = x509_ext
string_mask         = utf8only

[ ca ]
default_ca = ca_default

[ ca_default ]
new_certs_dir = .              # Location for new certs after signing
database      = ./index.txt    # Database index file
serial        = ./serial.txt   # The current serial number

default_days  = 1000
default_md    = sha256

policy        = signing_policy
email_in_dn   = no


[ intermediate_cert_extensions ]
keyUsage         = keyCertSign, digitalSignature
basicConstraints = CA:TRUE, pathlen:1

# 중간 인증서 생성 시 필수 설정
[ signing_policy ]
countryName            = optional
stateOrProvinceName    = optional
localityName           = optional
organizationName       = optional
organizationalUnitName = optional
commonName             = supplied
emailAddress           = optional

# 클라이언트용 설정
[client_cert_extensions]
keyUsage         = keyCertSign, digitalSignature
basicConstraints = CA:FALSE
subjectAltName      = @alternate_names


# The Subject DN can be formed using X501 or RFC 4514 (see RFC 4519 for a description).
#   Its sort of a mashup. For example, RFC 4514 does not provide emailAddress.
[ subject ]
countryName         = Country Name (2 letter code)
countryName_default = US

stateOrProvinceName         = State or Province Name (full name)
stateOrProvinceName_default = FL

localityName          = Locality Name (eg, city)
localityName_default  = Florida

organizationName         = Organization Name (eg, company)
organizationName_default = Andrew Connell Inc.

# Use a friendly name here because its presented to the user. The server's DNS
#   names are placed in Subject Alternate Names. Plus, DNS names here is deprecated
#   by both IETF and CA/Browser Forums. If you place a DNS name here, then you
#   must include the DNS name in the SAN too (otherwise, Chrome and others that
#   strictly follow the CA/Browser Baseline Requirements will fail).
commonName          = Common Name (e.g. server FQDN or YOUR name)
commonName_default  = localhost

emailAddress         = Email Address
emailAddress_default = brickwall@andrewconnell.com

# Section x509_ext is used when generating a self-signed certificate. I.e., openssl req -x509 ...
[ x509_ext ]

subjectKeyIdentifier    = hash
authorityKeyIdentifier  = keyid,issuer

# You only need digitalSignature below. *If* you don't allow
#   RSA Key transport (i.e., you use ephemeral cipher suites), then
#   omit keyEncipherment because that's key transport.
basicConstraints  = CA:FALSE
keyUsage          = digitalSignature, keyEncipherment
# 여기확인
subjectAltName    = @alternate_names
nsComment         = "OpenSSL Generated Certificate"

# RFC 5280, Section 4.2.1.12 makes EKU optional
#   CA/Browser Baseline Requirements, Appendix (B)(3)(G) makes me confused
#   In either case, you probably only need serverAuth.
# extendedKeyUsage  = serverAuth, clientAuth

# Section req_ext is used when generating a certificate signing request. I.e., openssl req ...
[ req_ext ]

subjectKeyIdentifier        = hash

basicConstraints    = CA:FALSE
keyUsage            = digitalSignature, keyEncipherment
# 여기확인
subjectAltName      = @alternate_names
nsComment           = "OpenSSL Generated Certificate"

# RFC 5280, Section 4.2.1.12 makes EKU optional
#   CA/Browser Baseline Requirements, Appendix (B)(3)(G) makes me confused
#   In either case, you probably only need serverAuth.
# extendedKeyUsage  = serverAuth, clientAuth

[ alternate_names ]

DNS.1       = localhost
DNS.2       = localhost.localdomain
DNS.3       = 127.0.0.1

# DNS.1       = example.com
# DNS.2       = www.example.com
# DNS.3       = mail.example.com
# DNS.4       = ftp.example.com

# Add these if you need them. But usually you don't want them or
#   need them in production. You may need them for development.
# DNS.5       = localhost
# DNS.6       = localhost.localdomain
# DNS.7       = 127.0.0.1

# IPv6 localhost
# DNS.8     = ::1

EOF
`

	three := `echo '01' > serial.txt && touch index.txt`
	four := `openssl req -x509 -nodes -newkey rsa:1024 -config rootca.conf -keyout rootca.key -out rootca.crt`
	five := `openssl req -nodes -newkey rsa:2048 -subj '/CN=NO.1 Certificate' -keyout intermediate.key -out intermediate.csr`
	six := `yes | openssl ca -config serverca.conf -extensions intermediate_cert_extensions -cert rootca.crt -keyfile rootca.key -out intermediate.crt -infiles intermediate.csr`
	seven := `openssl x509 -text -in intermediate.crt`
	eight := `openssl req -nodes -newkey rsa:2048 -subj '/CN=localhost' -keyout server.key -out server.csr`
	nine := `yes | openssl ca -config serverca.conf -extensions client_cert_extensions -cert intermediate.crt -keyfile intermediate.key -out server.crt -infiles server.csr`
	ten := `openssl x509 -text -in server.crt`

	writeAndRunScript("rootca.conf", one)
	writeAndRunScript("serverca.conf", two)
	executeShellCommand(three)
	executeShellCommand(four)
	executeShellCommand(five)
	executeShellCommand(six)
	executeShellCommand(seven)
	executeShellCommand(eight)
	executeShellCommand(nine)
	executeShellCommand(ten)

}

func MakrDirectory() {
	tmpDir = "tmp"
	err := os.Mkdir(tmpDir, 0755)
	if err != nil {
		panic(err)
	}

	fmt.Println("Temporary directory created:", tmpDir)
}

func ChangeDirectory() {
	err := os.Chdir(tmpDir)
	if err != nil {
		panic(err)
	}

	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}
	fmt.Println("Current working directory:", currentDir)
}

func writeAndRunScript(filename, content string) {
	// 스크립트를 파일로 작성합니다.
	// filepath := tmpDir + "/" + filename

	err := os.WriteFile(filename, []byte(content), 0755)
	if err != nil {
		panic(err) // 오류 발생 시 패닉
	}

	// 작성된 스크립트를 실행합니다.
	cmd := exec.Command("bash", filename)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		panic(err) // 오류 발생 시 패닉
	}
}

func executeShellCommand(command string) {
	// 명령어를 쉘에서 실행하도록 설정
	execCmd := exec.Command("sh", "-c", command)

	// 명령어 실행
	err := execCmd.Run()
	if err != nil {
		panic(err) // 오류 발생 시 패닉
	}
}
