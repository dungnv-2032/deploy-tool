package deploy

import (
	"fmt"
	"github.com/dung13890/deploy-tool/cmd/task"
	"github.com/dung13890/deploy-tool/utils"
	"sync"
	// "strings"
)

type Cluster struct {
	hosts []string
	rsync struct {
		excludes []string
	}
	cmds []string
}

func NewCluster(hosts []string, excludes []string, cmds []string) *Cluster {
	c := &Cluster{
		hosts: hosts,
		cmds:  cmds,
	}
	c.rsync.excludes = excludes

	return c
}

func (c *Cluster) Run(t *task.Task) error {
	// if hosts is empty then return
	if len(c.hosts) == 0 {
		return nil
	}
	// Make command rsync with excludes
	cmdRsync := "rsync -ahv --delete --omit-dir-times"
	if len(c.rsync.excludes) > 0 {
		excludes := " --excludes='*/releases/*'"
		for _, v := range utils.UniqueArr(c.rsync.excludes) {
			excludes += fmt.Sprintf(" --excludes='%s'", v)
		}
		cmdRsync += excludes
	}
	// Run Rsync
	if err := c.cmdRsync(t, cmdRsync); err != nil {
		return err
	}

	// Run Command
	for _, v := range c.cmds {
		if v != "" {
			c.command(t, v)
		}
	}

	return nil
}

func (c *Cluster) cmdRsync(t *task.Task, cmdRsync string) error {
	path := t.Dir()
	// Use go routine for unique clients
	hosts := utils.UniqueArr(c.hosts)
	wg := sync.WaitGroup{}
	errCh := make(chan error, len(hosts))
	for _, host := range hosts {
		wg.Add(1)
		go func(t *task.Task, host string, path string, cmd string) {
			defer wg.Done()
			cmd = fmt.Sprintf("%s %s %s:%s", cmd, path, host, path)
			if err := t.Run(cmd); err != nil {
				errCh <- err
			}

		}(t, host, path, cmdRsync)
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		return err
	}

	return nil
}

func (c *Cluster) command(t *task.Task, cmd string) error {
	path := t.Dir()
	releasePath := fmt.Sprintf("%s/release", path)
	// Use go routine for unique clients
	hosts := utils.UniqueArr(c.hosts)
	wg := sync.WaitGroup{}
	errCh := make(chan error, len(hosts))

	for _, host := range hosts {
		wg.Add(1)
		go func(t *task.Task, host string, path string, cmd string) {
			defer wg.Done()
			cmd = fmt.Sprintf("ssh %s cd %s && %s", host, releasePath, cmd)
			if err := t.Run(cmd); err != nil {
				errCh <- err
			}

		}(t, host, path, cmd)
	}
	wg.Wait()
	close(errCh)

	return nil
}
