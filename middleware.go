package goravel

import "net/http"

func (grvl *Goravel) SessionLoad(next http.Handler) http.Handler {
	grvl.InfoLog.Println("SessionLoad Called")
	return grvl.Session.LoadAndSave(next)
}
