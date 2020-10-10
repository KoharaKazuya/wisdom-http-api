このコンポーネントは信頼できる認証機能付きプロキシのバックエンドとして動かすことを想定する。
そのため、アクセスされた時点で無条件に識別情報 (= `wisdom-cognito-username` に設定されたメールアドレス) を信用し、認可機能は持たない。

ユーザーデータの提供元として Cognito User Pool に依存しているため、このスタックのパラメーターとして Cognito User Pool ID を要求する。このため `wisdom-cloud` とは相互依存となっており、作成の順序の問題が発生する。
`wisdom-http-api` (Cognito User Pool ダミードメインの使用、HTTP API ドメインの発行) → `wisdom-cloud` (HTTP API ドメインの使用、Cognito User Pool ドメインの発行) → `wisdom-http-api` (Cognito User Pool ドメインの使用) の順番で設定すると解決できる。

**前提**

- SSH 鍵ペアを作成する
  - パスフレーズは空文字列
- コンテンツリポジトリ (= GitHub 上の `wisdom-content`) に **書き込み権限付きの** デプロイキーとして公開鍵を登録する
- 秘密鍵を System Manager パラメータストアに保存する
  - キーの名前: `/wisdom/wisdom-http-api/deploy-key`
  - 種類: `SecureString`
  - AWS CLI で `aws ssm put-parameter --name /wisdom/wisdom-http-api/deploy-key --description 'Wisdom HTTP API からコンテンツリポジトリに読み書きするための秘密鍵' --value <秘密鍵のファイル内容> --type SecureString` など

**デプロイ**

- `make deploy`
