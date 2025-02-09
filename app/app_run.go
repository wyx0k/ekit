package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (r *RootComponent) runAll() error {
	var errs []error
	idx := 0
	ctx, cancel := context.WithCancel(context.Background())
	for _, c := range r.app.components {
		if rc, ok := c.component.(RunnableComponent); ok {
			errs = append(errs, nil)
			r.runningWg.Add(1)
			go func(i int) {
				defer func() {
					if err := recover(); err != nil {
						errs[i] = fmt.Errorf("component[%s] panic at running:%v", c.ID(), err)
					}
					r.runningWg.Done()
				}()
				err := rc.Run(r.app, r.conf)
				if err != nil {
					errs[i] = err
				}
			}(idx)
			idx++
		}
	}
	go func() {
		r.runningWg.Wait()
		cancel()
	}()
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
	EXIT:
		for {
			sig := <-sc
			r.logger.Infof("got signal: %v\n", sig)
			switch sig {
			case syscall.SIGQUIT, syscall.SIGINT:
				r.app.Exit("app canceled by user, stopping...")
			case syscall.SIGTERM:
				cancel()
				break EXIT
			default:
			}
		}
	}()
	select {
	case msg, ok := <-r.exitNotifyCh:
		if !ok {
			panic("system has been closed before")
		}
		r.logger.Info("app exit:", msg)
		if r.gracefulShutdownTimeout == 0 {
			<-r.exitFinishedCh
		} else {
			select {
			case <-r.exitFinishedCh:
			case <-time.After(r.gracefulShutdownTimeout):
			}
		}
		if len(r.app.exitErrs) > 0 {
			r.logger.Error(errors.Join(r.app.exitErrs...))
		}
	case <-ctx.Done():
		if len(r.app.exitErrs) > 0 {
			r.logger.Error(errors.Join(r.app.exitErrs...))
		}
	}
	return errors.Join(errs...)
}
