package sync

import (
	"path"
	"sort"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
	log "github.com/sirupsen/logrus"
	errors "golang.org/x/xerrors"

	"github.com/raystlin/ftpsync/config"
)

func Sync(conf *config.Config) error {

	source, err := newClient(&conf.Source)
	if err != nil {
		log.WithFields(log.Fields{
			"error":    err,
			"server":   conf.Source.Server,
			"username": conf.Source.Username,
		}).Error("Login error")
		return errors.Errorf("Error creating source connection %v", err)
	}
	defer source.Quit()

	if strings.HasPrefix(conf.Source.Path, "*") {
		log.WithFields(log.Fields{
			"path":   conf.Source.Path,
			"server": "source",
		}).Debug("Unknown real path, searching for it")

		conf.Source.Path, err = findPath(source, "/", strings.TrimPrefix(conf.Source.Path, "*"))
		if err != nil {
			return errors.Errorf("Could not find the source path: %v", err)
		}

		log.WithFields(log.Fields{
			"path":   conf.Source.Path,
			"server": "source",
		}).Info("Unknown path resolved")
	}

	dest, err := newClient(&conf.Dest)
	if err != nil {
		log.WithFields(log.Fields{
			"error":    err,
			"server":   conf.Dest.Server,
			"username": conf.Dest.Username,
		}).Error("Error connecting to server")
		return errors.Errorf("Error creating dest connection %v", err)
	}
	defer dest.Quit()

	if strings.HasPrefix(conf.Dest.Path, "*") {
		log.WithFields(log.Fields{
			"path":   conf.Dest.Path,
			"server": "dest",
		}).Debug("Unknown real path, searching for it")

		conf.Dest.Path, err = findPath(dest, "/", strings.TrimPrefix(conf.Dest.Path, "*"))
		if err != nil {
			return errors.Errorf("Could not find the dest path: %v", err)
		}

		log.WithFields(log.Fields{
			"path":   conf.Dest.Path,
			"server": "dest",
		}).Info("Unknown path resolved")
	}

	toDo := make(chan Job, 10*conf.NumClients)

	scheduler := NewScheduler(toDo)
	for i := 0; i < conf.NumClients; i++ {
		src, err := newClient(&conf.Source)
		if err != nil {
			log.WithFields(log.Fields{
				"error":      err,
				"thread-num": i,
				"server":     conf.Source.Server,
				"username":   conf.Source.Username,
			}).Error("Thread source client error")
			return errors.Errorf("Error creating thread source client: %v", err)
		}
		defer src.Quit()

		dst, err := newClient(&conf.Dest)
		if err != nil {
			log.WithFields(log.Fields{
				"error":      err,
				"thread-num": i,
				"server":     conf.Source.Server,
				"username":   conf.Source.Username,
			}).Error("Thread dest client error")
			return errors.Errorf("Error creating thread dest client: %v", err)
		}
		defer dst.Quit()

		scheduler.AddFetcher(src, dst)
	}

	err = checkPath(conf.Source.Path, conf.Dest.Path, toDo, source, dest)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Could not check the paths")
		return err
	}

	log.WithFields(log.Fields{
		"numElements": len(toDo),
	}).Info("Enum finished, syncing")

	close(toDo)
	scheduler.Wait()

	return nil
}

func newClient(conf *config.ServerConfig) (*ftp.ServerConn, error) {

	conn, err := ftp.Dial(conf.Server, ftp.DialWithTimeout(2*time.Minute), ftp.DialWithDisabledEPSV(true))
	if err != nil {
		return nil, err
	}

	err = conn.Login(conf.Username, conf.Password)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func findPath(conn *ftp.ServerConn, prevPath, searchedPath string) (string, error) {
	list, err := conn.List(prevPath)
	if err != nil {
		return "", err
	}

	for _, entry := range list {
		candidatePath := path.Join(prevPath, entry.Name, searchedPath)
		_, err := conn.List(candidatePath)
		if err == nil {
			return candidatePath, nil
		}
	}

	for _, entry := range list {
		candidatePath := path.Join(prevPath, entry.Name)
		realPath, err := findPath(conn, candidatePath, searchedPath)
		if err == nil {
			return realPath, nil
		}
	}

	return "", errors.New("Could not find the path")
}

func checkPath(sourcePath, destPath string, toDo chan<- Job, source, dest *ftp.ServerConn) error {

	sourceList, err := source.List(sourcePath)
	if err != nil {
		return errors.Errorf("[checkPath]: Error listing source: %v", err)
	}

	destList, err := dest.List(destPath)
	if err != nil {
		return errors.Errorf("[checkPath]: Error listing dest: %v", err)
	}

	byName := ByName(destList)
	sort.Sort(byName)

	for _, entry := range sourceList {
		if entry.Type == ftp.EntryTypeLink {
			continue
		}

		sourceFullPath := path.Join(sourcePath, entry.Name)
		destFullPath := path.Join(destPath, entry.Name)

		second := byName.Search(entry.Name)
		switch entry.Type {
		case ftp.EntryTypeFile:

			if second == nil || !AreEqual(entry, second) {
				toDo <- Job{
					SourcePath: sourceFullPath,
					DestPath:   destFullPath,
					Type:       entry.Type,
				}
			}
		case ftp.EntryTypeFolder:
			if second == nil {
				err := dest.MakeDir(destFullPath)
				if err != nil {
					return errors.Errorf("[checkPath]: Error making dest dir")
				}
			}
			checkPath(sourceFullPath, destFullPath, toDo, source, dest)
		}
	}

	return nil
}
