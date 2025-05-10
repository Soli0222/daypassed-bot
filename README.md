# DaysPassed Bot

指定された日付から経過した日数を計算し、Misskeyに投稿するBotです。

## 概要

このBotは、設定された特定の日付から現在までの経過日数を計算し、その結果をMisskeyのノートとして投稿します。KubernetesのCronJobとして定期実行されることを想定して設計されており、Dockerイメージとしてパッケージングされています。

## 主な機能

*   指定された日付からの経過日数を計算
*   計算結果をMisskeyに投稿
*   タイムゾーンの設定に対応
*   Misskeyインスタンスのホスト名とアクセストークンを設定可能
*   （オプション）1Passwordと連携し、Misskeyのアクセストークンを安全に取得
*   Dockerイメージとしてビルド可能
*   Kubernetesへのデプロイ用のHelmチャートを提供

## 構成

*   **Bot本体**: Go言語 (`main.go`)
*   **コンテナ化**: Docker (`Dockerfile`, `docker-compose.yaml`)
*   **Kubernetesデプロイ**: Helm (`charts/daypassed-bot/`)
*   **CI/CD**: GitHub Actions (`.github/workflows/docker.yaml`)

## 必要なもの

*   Docker
*   Go (開発時)
*   Kubernetesクラスタ (Helmチャートでデプロイする場合)
*   Misskeyアカウントとアクセストークン

## セットアップと実行

### 1. Dockerイメージのビルド

リポジトリのルートで以下のコマンドを実行してDockerイメージをビルドします。

```bash
docker build -t daypassed-bot .
```

GitHub Actionsにより、タグがプッシュされると自動的に`ghcr.io/${{ github.repository }}`にイメージがビルド・プッシュされます。

### 2. ローカルでの実行 (Docker Compose)

`docker-compose.yaml` を編集して、必要な環境変数を設定します。

*   `SPECIFIC_DATE`: 経過日数を計算する基準日 (例: `2002-02-22`)
*   `MK_TOKEN`: Misskeyのアクセストークン
*   `MISSKEY_HOST`: Misskeyインスタンスのホスト名 (例: `example.tld`)
*   `TZ`: タイムゾーン (例: `Asia/Tokyo`)

設定後、以下のコマンドで実行できます。

```bash
docker compose run --rm app
```

### 3. Kubernetesへのデプロイ (Helm)

Helmチャートを使用してKubernetesクラスタにデプロイできます。

1.  `charts/daypassed-bot/values.yaml` を環境に合わせて編集します。主な設定項目は以下の通りです。
    *   `image.repository`: 使用するDockerイメージのリポジトリ (デフォルト: `ghcr.io/soli0222/daypassed-bot`)
    *   `image.tag`: 使用するDockerイメージのタグ
    *   `schedule`: CronJobの実行スケジュール (デフォルト: `0 0 * * *` - 毎日0時0分)
    *   `env.specificDate`: 経過日数を計算する基準日
    *   `env.misskeyHost`: Misskeyインスタンスのホスト名
    *   `env.tz`: タイムゾーン
    *   `onepassword.enabled`: 1Password連携を有効にするか (デフォルト: `false`)
        *   有効にする場合、`onepassword.itemPath` と `onepassword.tokenFieldInItem` を設定します。
    *   `existingSecret`: 1Password連携を使用しない場合、Misskeyトークンを格納した既存のKubernetes Secret名を指定します。
        *   `secretTokenKey` でSecret内のキー名を指定します。

    デプロイ環境に合わせた設定ファイル (例: `values.polester.yaml`) を作成し、それを利用することも推奨されます。

2.  Helmコマンドでデプロイします。

    ```bash
    # values.yaml を使用する場合
    helm install <リリース名> ./charts/daypassed-bot

    # カスタム設定ファイルを使用する場合 (例: values.polester.yaml)
    helm install <リリース名> ./charts/daypassed-bot -f ./charts/daypassed-bot/values.polester.yaml
    ```

## 設定項目

### 環境変数 (Bot本体)

Bot (`main.go`) は以下の環境変数を参照します。

*   `SPECIFIC_DATE`: (必須) 経過日数を計算する基準日 (フォーマット: `YYYY-MM-DD`)
*   `MK_TOKEN`: (必須) Misskeyのアクセストークン
*   `MISSKEY_HOST`: (必須) Misskeyインスタンスのホスト名
*   `TZ`: タイムゾーン (デフォルト: `Asia/Tokyo`)

### Helmチャート (`values.yaml`)

Helmチャートでは、上記環境変数に加え、CronJobのスケジュール、リソース制限、1Password連携などの設定が可能です。詳細は `charts/daypassed-bot/values.yaml` を参照してください。

## CI/CD

GitHub Actions (`.github/workflows/docker.yaml`) により、以下の処理が自動化されています。

*   Gitのタグがプッシュされた際にトリガー
*   Dockerイメージをビルド (Linux/AMD64, Linux/ARM64)
*   ビルドされたイメージをGitHub Container Registry (`ghcr.io`) にプッシュ

## 開発

### Goプログラムの実行

ローカルでGoプログラムを実行する場合は、必要な環境変数を設定した上で `main.go` を実行します。

```bash
export SPECIFIC_DATE="2002-02-22"
export MK_TOKEN="your_misskey_token"
export MISSKEY_HOST="your_misskey_host"
export TZ="Asia/Tokyo"
go run main.go
```

## 注意事項

*   Misskeyのアクセストークン (`MK_TOKEN`) は機密情報です。1Password連携やKubernetes Secretを利用して安全に管理してください。
*   `SPECIFIC_DATE` は `YYYY-MM-DD` の形式で指定してください。
*   `MISSKEY_HOST` にはプロトコル (`https://`) を含めないでください。

