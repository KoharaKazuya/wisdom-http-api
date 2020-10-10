このコンポーネントは信頼できる認証機能付きプロキシのバックエンドとして動かすことを想定する。
そのため、アクセスされた時点で無条件にユーザー情報 (= `wisdom-user-name`, `wisdom-user-email` ヘッダー) を信用し、認可機能は持たない。

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
