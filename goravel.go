package goravel

const version = "1.0.0"

type Goravel struct {
	AppName string
	Debug   bool
	Version string
}

func (grvl *Goravel) New(rootPath string) error {
	pathConfig := initPaths{
		rootPath:    rootPath,
		folderNames: []string{"handlers", "migrations", "views", "data", "public", "tmp", "logs", "middleware"},
	}

	err := grvl.Init(pathConfig)
	if err != nil {
		return err
	}

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
