// This file is part of autosr.
//
// autosr is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// autosr is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with autosr.  If not, see <https://www.gnu.org/licenses/>.

package track

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/bobbytrapz/autosr/options"
)

// when a stream appears to end we wait to see if the user comes back
// online just in case there was a problem with the stream
var recoverTimeout = 5 * time.Minute

type saveTask struct {
	name string
	link string
}

var saving = struct {
	sync.RWMutex
	tasks map[saveTask]time.Time
}{
	tasks: make(map[saveTask]time.Time),
}

// find the most recent save for a link
func findSaveTask(link string) (task saveTask, createdAt time.Time) {
	saving.RLock()
	defer saving.RUnlock()

	// find the most recently added save task matching our target
	for t, at := range saving.tasks {
		if t.link == link {
			if createdAt.IsZero() || at.After(createdAt) {
				createdAt = at
				task = t
			}
		}
	}

	return
}

func hasSaveTask(task saveTask) bool {
	saving.RLock()
	defer saving.RUnlock()
	_, ok := saving.tasks[task]
	return ok
}

// give true if it is newly added
func addSaveTask(task saveTask) bool {
	if hasSaveTask(task) {
		return false
	}
	saving.Lock()
	defer saving.Unlock()
	saving.tasks[task] = time.Now()
	return true
}

func delSaveTask(task saveTask) {
	saving.Lock()
	defer saving.Unlock()
	delete(saving.tasks, task)
}

// record stream to disk using external program
func performSave(ctx context.Context, t *tracked, streamURL string) error {
	wg.Add(1)
	defer wg.Done()

	name := t.Name()

	link := t.Link()
	if link == "" {
		return errors.New("track.save: no link")
	}

	if streamURL == "" {
		return errors.New("track.save: no stream url")
	}

	task := saveTask{
		name: t.Name(),
		link: t.Link(),
	}
	if !addSaveTask(task) {
		log.Println("track.save: already saving", task.name)
		return nil
	}
	defer func() {
		delSaveTask(task)
		t.EndSave()
	}()
	t.BeginSave()
	log.Println("track.save:", task.name)

	// used by command monitor to indicate that the command has exited
	exit := make(chan error, 1)

	// command information set in the closure below
	var cmd *exec.Cmd
	var app string
	var pid int

	// will be called again if we manage to recover a stream
	runSave := func(url string) error {
		var err error
		cmd, err = runDownloader(ctx, url, name)
		if err != nil {
			return fmt.Errorf("track.save: %s", err)
		}

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("track.save: %s", err)
		}
		app = cmd.Args[0]
		pid = cmd.Process.Pid
		log.Printf("track.save: %s [%s %d]", name, app, pid)

		// monitor downloader
		go func() {
			err := cmd.Wait()
			exit <- err
		}()

		return nil
	}

	// try to run the save command now
	if err := runSave(streamURL); err != nil {
		return err
	}

	cancelSave := make(chan struct{})
	t.SetCancel(cancelSave)

	// handle closing downloader
	for {
		select {
		case <-ctx.Done():
			cmd.Process.Kill()
			err := cmd.Wait()
			t.SetFinishedAt(time.Now())
			log.Printf("track.save: %s %s [%s %d] (%s)", name, ctx.Err(), app, pid, err)
			return nil
		case <-cancelSave:
			// we have been selected for cancellation
			cmd.Process.Kill()
			err := cmd.Wait()
			t.SetFinishedAt(time.Now())
			log.Printf("track.save: %s canceled [%s %d] (%s)", name, app, pid, err)
			return nil
		case err := <-exit:
			if err != nil {
				// something may have gone wrong so try to recover
				log.Printf("track.save: %s exited [%s %d]", name, app, pid)
				d, newURL, err := maybeRecover(ctx, t)
				if err != nil {
					// we did not recover so end this save
					t.SetFinishedAt(time.Now().Add(-d))
					return nil
				}
				log.Printf("track.save: %s recovered (%s)", name, d.Truncate(time.Millisecond))
				// run a new save command
				runSave(newURL)
			} else {
				log.Printf("track.save: %s exit ok [%s %d]", name, app, pid)
				t.SetFinishedAt(time.Now())
				return nil
			}
		}
	}
}

// runs the user's downloader
func runDownloader(ctx context.Context, url, name string) (cmd *exec.Cmd, err error) {
	saveTo := filepath.Join(options.Get("save_to"), name)
	ua := fmt.Sprintf("User-Agent=%s", options.Get("user_agent"))

	fn := fmt.Sprintf("%s-%s", time.Now().Format("2006-01-02"), name)
	saveAs := fn
	for n := 2; ; n++ {
		p := filepath.Join(saveTo, saveAs+".ts")
		if _, err := os.Stat(p); os.IsNotExist(err) {
			break
		}
		saveAs = fmt.Sprintf("%s %d", fn, n)
	}
	saveAs = saveAs + ".ts"

	app := options.Get("download_with")
	args := []string{
		"--http-header", ua,
		"-o", saveAs,
		fmt.Sprintf("hlsvariant://%s", url),
		"best",
	}

	cmd = exec.CommandContext(ctx, app, args...)
	err = os.MkdirAll(saveTo, os.ModePerm)
	if err != nil {
		err = fmt.Errorf("track.RunDownloader: %s", err)
		return
	}
	cmd.Dir = saveTo
	setArgs(cmd)

	return cmd, nil
}

func maybeRecover(ctx context.Context, t *tracked) (duration time.Duration, streamURL string, err error) {
	beginAt := time.Now()
	defer func() {
		endAt := time.Now()
		duration = endAt.Sub(beginAt)
	}()

	name := t.Name()
	log.Println("track.maybeRecover:", name, "begin")

	err = waitForLive(ctx, t, recoverTimeout)
	if err != nil {
		log.Println("track.maybeRecover:", name, "is not online")
		err = errors.New("track.maybeRecover: target is not live")
		return
	}
	log.Println("track.maybeRecover:", name, "is online")

	streamURL, err = waitForStream(ctx, t, recoverTimeout)
	if err != nil {
		// we failed to find the new url
		log.Println("track.maybeRecover:", name, "did not find url")
		err = errors.New("track.maybeRecover: did not find url")
		return
	}

	log.Println("track.maybeRecover:", name, "found url")

	return
}
