참고

https://www.voitanos.io/blog/updated-creating-and-trusting-self-signed-certs-on-macos-and-chrome/

server 인증서에는 san이 필수

크롬 개발자 도구 > security 에서 확인 가능

```sh
cat server.crt intermediate.crt > server_with_chain.crt
```
이걸 해줘야 certificate chaining이 됨
