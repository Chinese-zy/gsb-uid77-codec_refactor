在仓库根目录执行：

go run ./cmd/api

然后打开 http://127.0.0.1:8777
不要装包，也没有容器。

页面和接口共用同一套报文字节规矩，见 `SPEC.md`：
小端多字节整数 + LEB128 变长整数（看最高位续位），字段严格按出现顺序
数组保留、重复不去重。Go 实现在 `frame/codec.go`，页面实现在
`web/codec.js`，两边对同一份字节必须吐出相同结果。

测试（写死的字节对照表，Go 测试会直接调 Node 对拍）：

go test ./...
node --test web/codec.test.js
