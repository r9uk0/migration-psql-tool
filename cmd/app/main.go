package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	model "migration-psql-tool/internal/domain/models"
	zlog "migration-psql-tool/pkg/logger"

	"github.com/joho/godotenv"
)

type SelectEnvInfo struct {
	SelectHost string `json:"SelectHost"`
	SelectPort string `json:"SelectPort"`
	SelectUser string `json:"SelectUser"`
	SelectName string `json:"SelectName"`
	SelectPass string `json:"SelectPass"`
}

type InsertEnvInfo struct {
	InsertHost string `json:"InsertHost"`
	InsertPort string `json:"InsertPort"`
	InsertUser string `json:"InsertUser"`
	InsertName string `json:"InsertName"`
	InsertPass string `json:"InsertPass"`
}

func run() error {
	// ログ初期化
	zlog.Init(os.Stdout)

	// テーブル名取得
	tables := model.NewTableNames()

	// 環境変数読み込み
	err := godotenv.Load("../../.env")
	if err != nil {
		zlog.Error().Msgf(".envの読み込み失敗 from run() %v", err)
	}

	// 引数によって実行する関数を決める
	const minArgsLen = 2
	if len(os.Args) < minArgsLen {
		zlog.Error().Msg("引数が不足しています。実行する関数を指定してください。")
		return nil
	}
	const firstArg = 1
	switch os.Args[firstArg] {
	case "selectToDump":
		// .env ファイルから環境変数を取得
		selectEnvInfo := SelectEnvInfo{
			SelectHost: os.Getenv("SELECT_DB_HOST"),
			SelectPort: os.Getenv("SELECT_DB_PORT"),
			SelectUser: os.Getenv("SELECT_DB_USER"),
			SelectName: os.Getenv("SELECT_DB_NAME"),
			SelectPass: os.Getenv("SELECT_DB_PASSWORD"),
		}
		zlog.Info().Msgf("selectEnvInfo %v", selectEnvInfo)

		// テーブル名分ループ
		for _, tableName := range tables.Tables {
			// dump出力
			err := selectToDump(tableName, selectEnvInfo)
			if err != nil {
				zlog.Error().Msgf("dump出力失敗 from run() %v", err)
				return err
			}
			zlog.Info().Msg("dump出力成功 from run()")
		}
	case "dumpToInsert":
		// .env ファイルから環境変数を取得
		InsertEnvInfo := InsertEnvInfo{
			InsertHost: os.Getenv("INSERT_DB_HOST"),
			InsertPort: os.Getenv("INSERT_DB_PORT"),
			InsertUser: os.Getenv("INSERT_DB_USER"),
			InsertName: os.Getenv("INSERT_DB_NAME"),
			InsertPass: os.Getenv("INSERT_DB_PASSWORD"),
		}
		zlog.Info().Msgf("InsertEnvInfo %v", InsertEnvInfo)

		// テーブル名分ループ
		for _, tableName := range tables.Tables {
			// dump出力
			err := dumpToInsert(tableName, InsertEnvInfo)
			if err != nil {
				zlog.Error().Msgf("dumpInsert失敗 from run() %v", err)
				return err
			}
			zlog.Info().Msg("dumpInsert成功 from run()")
		}
	default:
		zlog.Error().Msg("未知の関数です。")
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		zlog.Error().Msgf("err %v", err)
	}
}

func selectToDump(tableName string, selectEnvInfo SelectEnvInfo) error {
	// PostgreSQL接続情報
	// エクスポートするテーブル名
	zlog.Info().Msgf("dump出力開始 %v", tableName)

	// sqlファイル存在確認あればそのファイルを使う
	var selectQuery string
	var err error
	fileName := "../../internal/infrastructure/db/query/" + tableName + ".sql"
	isExists := fileExists(fileName)
	if isExists {
		selectQuery, err = readFileContents(fileName)
		if err != nil {
			zlog.Error().Msgf("SQLファイル読み込み err %v", err)
			return err
		}
	} else {
		// 特定のSELECT文がないならすべて取得
		selectQuery = "SELECT * FROM " + tableName
	}

	// COPYコマンドを作成
	query := fmt.Sprintf("COPY (%s) TO STDOUT WITH CSV HEADER", selectQuery)

	// 出力ファイルの作成
	outFile, err := os.Create(fmt.Sprintf("%s.dump", tableName))
	if err != nil {
		zlog.Error().Msgf("出力ファイルの作成 err %v", err)
		zlog.Error().Msgf("dump出力失敗 %v", tableName)
		return err
	}
	defer outFile.Close()

	// psqlコマンドを使ってCOPY TO STDOUTを実行し、結果をファイルに書き込む
	var cmd = exec.Command("psql", "-h", selectEnvInfo.SelectHost, "-p", selectEnvInfo.SelectPort, "-U", selectEnvInfo.SelectUser, "-d", selectEnvInfo.SelectName, "-c", query)
	cmd.Stdout = outFile
	cmd.Stderr = os.Stderr
	zlog.Info().Msgf("cmd %v", cmd)

	// 環境変数にパスワードを設定（必要な場合）
	cmd.Env = append(cmd.Env, "PGPASSWORD="+selectEnvInfo.SelectPass)

	// コマンド実行
	err = cmd.Run()
	if err != nil {
		zlog.Error().Msgf("dump出力失敗 %v", tableName)
		zlog.Error().Msgf("COPYコマンドの実行結果ファイル書き込みエラー %v", err)
		return err
	}

	zlog.Info().Msgf("dump出力成功 %v", tableName)
	return nil
}

func dumpToInsert(tableName string, insertEnvInfo InsertEnvInfo) error {
	// コマンドを実行してダンプを適用
	var psqlHead = "psql"

	// TRUNCATEコマンドを実行
	truncateCmd := exec.Command(psqlHead,
		"-h", insertEnvInfo.InsertHost,
		"-p", insertEnvInfo.InsertPort,
		"-U", insertEnvInfo.InsertUser,
		"-d", insertEnvInfo.InsertName,
		"-c", fmt.Sprintf("TRUNCATE TABLE %s", tableName))

	// psqlの環境変数にパスワードを追加
	truncateCmd.Env = append(os.Environ(), "PGPASSWORD="+insertEnvInfo.InsertPass)

	// 標準エラー出力をキャプチャ
	var truncateStderr bytes.Buffer
	truncateCmd.Stderr = &truncateStderr

	zlog.Info().Msg("TRUNCATEコマンドを実行")
	err := truncateCmd.Run()
	if err != nil {
		zlog.Error().Msgf("TRUNCATE失敗: %v", tableName)
		zlog.Error().Msgf("エラー: %v, stderr: %s", err, truncateStderr.String())
		return err
	}
	zlog.Info().Msgf("TRUNCATE成功 %v", tableName)

	// ダンプ適用コマンド
	cmd := exec.Command(psqlHead, "-h", insertEnvInfo.InsertHost, "-p", insertEnvInfo.InsertPort, "-U", insertEnvInfo.InsertUser, "-d", insertEnvInfo.InsertName, "-c", fmt.Sprintf("\\COPY "+tableName+" FROM '%s' WITH CSV HEADER", fmt.Sprintf("%s.dump", tableName)))
	cmd.Env = append(os.Environ(), "PGPASSWORD="+insertEnvInfo.InsertPass)
	zlog.Info().Msgf("cmd %v", cmd)

	// 標準エラー出力をキャプチャ
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// コマンド実行
	err = cmd.Run()
	if err != nil {
		zlog.Error().Msgf("dumpINSERT失敗 %v", tableName)
		zlog.Error().Msgf("dumpをテーブルにinsert失敗 err %v, stderr: %s", err, stderr.String())
		return err
	}

	zlog.Info().Msgf("dumpINSERT成功 %v", tableName)
	return nil
}

func fileExists(filename string) bool {
	// os.Stat はファイルまたはディレクトリの情報を取得
	_, err := os.Stat(filename)
	if err != nil {
		// エラーが返された場合、ファイルが存在しないか他の問題が発生した場合
		if os.IsNotExist(err) {
			// ファイルが存在しない場合
			return false
		}
		// その他のエラー（例: パーミッションエラーなど）
		fmt.Println("Error checking file:", err)
		return false
	}
	// ファイルが存在する場合
	return true
}

func readFileContents(filename string) (string, error) {
	// ioutil.ReadFileを使ってファイルを読み込む
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
