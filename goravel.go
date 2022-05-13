package goravel

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/CloudyKit/jet/v6"
	"github.com/NinaM31/goravel/render"
	"github.com/NinaM31/goravel/session"
	"github.com/alexedwards/scs/v2"
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
	Session  *scs.SessionManager
	DB       Database
	JetViews *jet.Set
	config   config
}

type config struct {
	port        string
	renderer    string
	cookie      cookieConfig
	sessionType string
	database    databaseConfig
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

	// connet to database
	if os.Getenv("DATABASE_TYPE") != "" {
		db, err := grvl.OpenDB(os.Getenv("DATABASE_TYPE"), grvl.BuildDSN())

		if err != nil {
			errorLog.Println(err)
			os.Exit(1)
		}

		grvl.DB = Database{
			DataType: os.Getenv("DATABASE_TYPE"),
			Pool:     db,
		}
	}

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
		cookie: cookieConfig{
			name:     os.Getenv("COOKIE_NAME"),
			lifetime: os.Getenv("COOKIE_LIFETIME"),
			persist:  os.Getenv("COOKIE_PERSISTS"),
			secure:   os.Getenv("COOKIE_SECURE"),
			domain:   os.Getenv("COOKIE_DOMAIN"),
		},
		sessionType: os.Getenv("SESSION_TYPE"),
		database: databaseConfig{
			database: os.Getenv("DATABASE_TYPE"),
			dsn:      grvl.BuildDSN(),
		},
	}

	// create session
	sess := session.Session{
		CookieLifetime: grvl.config.cookie.lifetime,
		CookiePersist:  grvl.config.cookie.persist,
		CookieName:     grvl.config.cookie.name,
		SessionType:    grvl.config.sessionType,
		CookieDomain:   grvl.config.cookie.domain,
	}
	grvl.Session = sess.InitSession()

	var views = jet.NewSet(
		jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
		jet.InDevelopmentMode(),
	)
	grvl.JetViews = views

	grvl.createRenderer()

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

	defer grvl.DB.Pool.Close()

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

func (grvl *Goravel) createRenderer() {
	myRenderer := render.Render{
		Renderer: grvl.config.renderer,
		RootPath: grvl.RootPath,
		Port:     grvl.config.port,
		JetViews: grvl.JetViews,
	}
	grvl.Render = &myRenderer
}

func (c *Goravel) BuildDSN() string {
	var dsn string

	switch os.Getenv("DATABASE_TYPE") {
	case "postgres", "postgresql":
		dsn = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s timezone=UTC connect_timeout=5",
			os.Getenv("DATABASE_HOST"),
			os.Getenv("DATABASE_PORT"),
			os.Getenv("DATABASE_USER"),
			os.Getenv("DATABASE_NAME"),
			os.Getenv("DATABASE_SSL_MODE"),
		)

		if os.Getenv("DATABASE_PASS") != "" {
			dsn = fmt.Sprintf("%s password=%s", dsn, os.Getenv("DATABASE_PASS"))
		}
	default:

	}

	return dsn
}
