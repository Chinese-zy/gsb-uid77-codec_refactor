在仓库根目录执行：

go run ./cmd/api

然后打开 http://127.0.0.1:8777
不要装包，也没有容器。

编解码规矩见 SPEC.md，页面（web/codec.js）和接口（frame/）都按它走。
两边共用同一份写死的字节对照表 testdata/golden.json，对不上就别交：

go test ./...
node web/codec_test.js
