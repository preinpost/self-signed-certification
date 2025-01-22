package script

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

var tmpDir string

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
default_bits            = 2048
default_md              = sha1
default_keyfile         = rootca.key
distinguished_name      = dn
req_extensions          = req_ext
#extensions 설정은 ROOT CA/중간 인증서/클라이언트 별로 따로 지정하므로 해당 옵션 주석
#extensions              = v3_ca

[ ca ]
default_ca = ca_default

[ ca_default ]
new_certs_dir = .              # Location for new certs after signing
database      = ./index.txt    # Database index file , 체인 인증서(중간 인증 및 클라이언트 인증서) 인덱싱 관리용
serial        = ./serial.txt   # The current serial number, 체인 인증서(중간 인증 및 클라이언트 인증서) 시리얼 번호 관리용
default_days  = 3650

default_md    = sha256

policy        = signing_policy
email_in_dn   = no

# default_crl_days: 새로운 CRL이 발급되는 주기를 일(day) 단위로 설정합니다. 예를 들어, 7로 설정하면 CRL이 7일마다 갱신됩니다.
# default_crl_hours: 새로운 CRL이 발급되는 주기를 시간(hour) 단위로 설정합니다. 예를 들어, 168로 설정하면 CRL이 168시간(7일)마다 갱신합니다.
default_crl_days = 7

# root CA용 설정
[ ca_cert_extensions ]
keyUsage               = keyCertSign, cRLSign #, digitalSignature
basicConstraints       = critical, CA:TRUE, pathlen:2
subjectKeyIdentifier   = hash
nsCertType             = sslCA, emailCA, objCA

# 중간 인증서용 설정
[ intermediate_cert_extensions ]
keyUsage         = keyCertSign, digitalSignature
basicConstraints = CA:TRUE, pathlen:1

# 클라이언트용 설정
[client_cert_extensions]
keyUsage         = keyCertSign, digitalSignature
basicConstraints = CA:FALSE

# 와일드 및 멀티 카드 인증서 설정
[ req_ext ]
subjectAltName = localhost
[ alt_names ]
DNS.1 = *.192.168.50.81.nip.io

# 중간 인증서 생성 시 필수 설정
[ signing_policy ]
countryName            = optional
stateOrProvinceName    = optional
localityName           = optional
organizationName       = optional
organizationalUnitName = optional
commonName             = supplied
emailAddress           = optional

# 기본 입력 값 또는 length 제한 설정
[ dn ]
# 발급자가 속한 국가명(C)
countryName                     = Country Name (2 letter code)
countryName_default             = KR  #
countryName_min                 = 2
countryName_max                 = 2

# 발급자가 속한 조직명(O)
organizationName              = Organization Name (eg, company)
organizationName_default      = Inje Inc.

# 발급자가 속한 하위 조직명(OU)
organizationalUnitName          = Organizational Unit Name (eg, section)
organizationalUnitName_default  = Cloud SW Team.

# 발급자명(CN)
commonName                      = Common Name (eg, your name or your servers hostname)
commonName_default              = Self Signed CA
commonName_max                  = 64

# 기타 값
ST=Seoul
L=Seoul
O=Inje Inc.
OU=Dev
emailAddress=owner@domain.co.kr

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
