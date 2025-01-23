### 체인 인증서 만들기

```sh
cat server.crt intermediate.crt > server_with_chain.crt
```
이걸 해줘야 certificate chaining이 됨


### 인증서 Windows 세팅

인증서 - 로컬 컴퓨터 에서
윈도우에서는 신뢰할 수 있는 루트 인증기관과 중간 인증 기관 둘다 넣어줘야함

rootca 또는 intermidiateca 는 신뢰할 수 있는 루트 인증기관에 넣어줌
실제 server.crt 는 중간 인증 기관에 넣어줘야함

chrome://restart

인증서 인식 이상하면 chrome 재부팅 해주기 (위 url을 복사해서 붙여넣어야함 그냥 껐다키면 안됨)

### 기타

참고 링크
+ https://www.voitanos.io/blog/updated-creating-and-trusting-self-signed-certs-on-macos-and-chrome/

server 인증서에는 san이 필수

크롬 개발자 도구 > security 에서 왜 인증서에 문제 있는지 확인 가능