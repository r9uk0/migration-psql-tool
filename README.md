# go-migration-psql-tool

## 目的

* DBデータをマスクして取得したい
  * マスクしながら取得するなら事前にSQLを用意しておく必要がある(両端をxに置換)
* 取得したデータを試験環境DBに流し込みたい

## 説明

### 環境構築

* Windows
  * [Goインストール](https://zenn.dev/atsushifx/articles/golang-win-install)
  * [PostgreSQL](https://qiita.com/tom-sato/items/037b8f8cb4b326710f71)
  * ちゃんと確認していなけれど以下も必要かも
    * ```export PGCLIENTENCODING=UTF8```
    * ```psql -h localhost -p 5432 -U postgres -d postgres```
    * ```SHOW client_encoding;```

* Linux(Ubuntu)
  * [Goインストール](https://www.server-world.info/query?os=Ubuntu_24.04&p=go&f=1#google_vignette)
  * [PostgreSQL](https://qiita.com/etaka/items/2c624275f090cc715f1e)

* Mac
  * [Goインストール](https://qiita.com/yoshihiro-kato/items/14e3fd701c63ee9a088a)
  * [PostgreSQL](https://zenn.dev/kento_kodama/articles/d67de221823eed)

### 実行手順

* app配下でのコマンド実行前提。
* PJ直下に以下のような.envを配置しておく。

```sh
ENV=local
# 取得元のDB情報
SELECT_DB_HOST=DBホスト名
SELECT_DB_PORT=DBポート番号
SELECT_DB_USER=DB接続ユーザー
SELECT_DB_NAME=DB名
SELECT_DB_PASSWORD=DB接続パスワード

# 生成先のDB情報
INSERT_DB_HOST=DBホスト名
INSERT_DB_PORT=DBポート番号
INSERT_DB_USER=DB接続ユーザー
INSERT_DB_NAME=DB名
INSERT_DB_PASSWORD=DB接続パスワード
```

#### 1.DBの最新データdump取得

```sh
go run main.go selectToDump
```

PJ直下に配置しておいた.envの情報(SELECT_XXXX)を読み込んで接続する。

#### 2.DBのテーブルをtruncateしコピー

```sh
go run main.go dumpToInsert
```

PJ直下に配置しておいた.envの情報(INSERT_XXXX)を読み込んで接続する。
スキーマ名などは適宜修正する。

## テーブル一覧

### XXスキーマ

* m_xxx

### 注意事項

環境差異のあるテーブルについては対象外とすること。

| テーブル物理名                   | テーブル論理名             | 調査内容                                                                                     | 移植可否 |
|----------------------------------|----------------------------|----------------------------------------------------------------------------------------------|--------------------|
| m_XXX                | XXXマスタ       | XXXのため                                                     | ×                 |