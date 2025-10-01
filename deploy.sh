#!/bin/sh
cd /Users/wellington/Developer/encerrar/encerrar_contrato/
flutter build web --release --dart-define=ENV=production --output /Users/wellington/Developer/encerrar/backend/public/ --base-href "/"
cd /Users/wellington/Developer/encerrar/backend/

ssh root@167.99.107.244 "pm2 stop all"
ssh root@167.99.107.244 "pm2 delete all"

ssh root@167.99.107.244 "rm -rf /root/public"

GOOS=linux GOARCH=amd64 go build -o build/main main.go
scp build/main root@167.99.107.244:/root

rm -rf public/assets/packages/flutter_multi_formatter/flags

scp -r public root@167.99.107.244:/root

ssh root@167.99.107.244 "pm2 start ./main --name encerrar_contrato"

