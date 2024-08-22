package routers

import (
	"bytes"
	"net/http"
	"os/exec"

	"github.com/google/shlex"
	"github.com/kajikentaro/elastic-proxy/models"
)

func NewHandleTemplate(handler models.Handler, contentType string, statusCode int, headers map[string]string) models.Handler {
	return &HandleTemplate{
		handler:     handler,
		contentType: contentType,
		statusCode:  statusCode,
		headers:     headers,
	}
}

type HandleTemplate struct {
	handler     models.Handler
	statusCode  int
	contentType string
	headers     map[string]string
}

// customResponseWriterは、http.ResponseWriterをラップして書き込まれたデータをキャプチャする
type customResponseWriter struct {
	http.ResponseWriter
	body bytes.Buffer
}

func (w *customResponseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func (h *HandleTemplate) Handle(w http.ResponseWriter, r *http.Request) {
	// customResponseWriterを使って、書き込まれたデータをキャプチャ
	crw := &customResponseWriter{ResponseWriter: w}

	// handlerの処理を実行し、その出力をキャプチャ
	h.handler.Handle(crw, r)

	INPUT := "sed -E 's/hoge/fuga/g'"
	cmdArgs, err := shlex.Split(INPUT)
	if err != nil {
		http.Error(w, "Failed to execute command", http.StatusInternalServerError)
		return
	}

	// コマンドを実行して標準入力にキャプチャしたボディを渡し、結果を取得
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdin = &crw.body

	// コマンドの実行結果を取得
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		http.Error(w, "Failed to execute command", http.StatusInternalServerError)
		return
	}

	// コンテンツタイプが指定されている場合、レスポンスヘッダーに設定
	if h.contentType != "" {
		w.Header().Set("Content-Type", h.contentType)
	}

	// ステータスコードが指定されている場合、レスポンスのステータスコードを設定
	if h.statusCode != 0 {
		w.WriteHeader(h.statusCode)
	}

	// その他のヘッダーを設定
	for v, k := range h.headers {
		w.Header().Set(v, k)
	}

	// コマンドの結果をレスポンスとして書き込む
	_, err = w.Write(out.Bytes())
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func (h *HandleTemplate) GetType() string {
	return h.handler.GetType()
}

func (h *HandleTemplate) GetResponseInfo() map[string]string {
	return h.handler.GetResponseInfo()
}
