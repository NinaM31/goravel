package goravel

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/NinaM31/goravel/render"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

const version = "1.0.0"

// Goravel is the type for the Goravel package
// members exported in this type is available to any application that uses it.
type Goravel struct {
	AppName  string
	Debug    bool
	Version  string
	ErrorLog *log.Logger
	InfoLog  *log.Logger
	RootPath string
	Routes   *chi.Mux
	Render   *render.Render
	config   config
}

type config struct {
	port     string
	renderer string
}

// New creates all the necessary folders, reads .env file,
// creates loggers and populates the Goravel type based on .env
func (grvl *Goravel) New(rootPath string) error {
	pathConfig := initPaths{
		rootPath:    rootPath,
		folderNames: []string{"handlers", "migrations", "views", "data", "public", "tmp", "logs", "middleware"},
	}

	err := grvl.Init(pathConfig)
	if err != nil {
		return err
	}

	err = grvl.checkDotEnv(rootPath)
	if err != nil {
		return err
	}

	// read .env
	err = godotenv.Load(rootPath + "/.env")
	if err != nil {
		return err
	}

	// create loggers
	infoLog, errorLog := grvl.startLoggers()

	// Populate grvl type
	grvl.InfoLog = infoLog
	grvl.ErrorLog = errorLog
	grvl.Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	grvl.AppName = os.Getenv("APP_NAME")
	grvl.Version = version
	grvl.RootPath = rootPath
	grvl.Routes = grvl.routes().(*chi.Mux)
	grvl.config = config{
		port:     os.Getenv("PORT"),
		renderer: os.Getenv("RENDERER"),
	}
	grvl.Render = grvl.createRenderer(grvl)

	return nil
}

func (grvl *Goravel) Init(p initPaths) error {
	root := p.rootPath
	for _, path := range p.folderNames {
		// create folder if doesn't exist
		err := grvl.CreateDirIfNoExist(root + "/" + path)
		if err != nil {
			return err
		}
	}
	return nil
}

// ListenAndServe starts the web server
func (grvl *Goravel) ListenAndServe() {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", os.Getenv("PORT")),
		ErrorLog:     grvl.ErrorLog,
		Handler:      grvl.Routes,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	grvl.InfoLog.Printf("Listening on port %s", os.Getenv("PORT"))
	err := srv.ListenAndServe()
	grvl.ErrorLog.Fatal(err)
}

func (grvl *Goravel) checkDotEnv(path string) error {
	err := grvl.CreateFileIfNoExists(fmt.Sprintf("%s/.env", path))
	if err != nil {
		return err
	}

	return nil
}

func (grvl *Goravel) startLoggers() (*log.Logger, *log.Logger) {
	var infoLog *log.Logger
	var errorLog *log.Logger

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	return infoLog, errorLog
}

func (grvl *Goravel) createRenderer(g *Goravel) *render.Render {
	myRenderer := render.Render{
		Renderer: g.config.renderer,
		RootPath: g.RootPath,
		Port:     g.config.port,
	}

	return &myRenderer
}
