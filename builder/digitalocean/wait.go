package digitalocean

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/digitalocean/godo"
)

// waitForDropletUnlocked waits for the Droplet to be unlocked to
// avoid "pending" errors when making state changes.
func waitForDropletUnlocked(
	client *godo.Client, dropletId int, timeout time.Duration) error {
	done := make(chan struct{})
	defer close(done)

	result := make(chan error, 1)
	go func() {
		attempts := 0
		for {
			attempts += 1

			log.Printf("[DEBUG] Checking droplet lock state... (attempt: %d)", attempts)
			droplet, _, err := client.Droplets.Get(context.TODO(), dropletId)
			if err != nil {
				result <- err
				return
			}

			if !droplet.Locked {
				result <- nil
				return
			}

			// Wait 3 seconds in between
			time.Sleep(3 * time.Second)

			// Verify we shouldn't exit
			select {
			case <-done:
				// We finished, so just exit the goroutine
				return
			default:
				// Keep going
			}
		}
	}()

	log.Printf("[DEBUG] Waiting for up to %d seconds for droplet to unlock", timeout/time.Second)
	select {
	case err := <-result:
		return err
	case <-time.After(timeout):
		return fmt.Errorf(
			"Timeout while waiting to for droplet to unlock")
	}
}

// waitForDropletState simply blocks until the droplet is in
// a state we expect, while eventually timing out.
func waitForDropletState(
	desiredState string, dropletId int,
	client *godo.Client, timeout time.Duration) error {
	done := make(chan struct{})
	defer close(done)

	result := make(chan error, 1)
	go func() {
		attempts := 0
		for {
			attempts += 1

			log.Printf("Checking droplet status... (attempt: %d)", attempts)
			droplet, _, err := client.Droplets.Get(context.TODO(), dropletId)
			if err != nil {
				result <- err
				return
			}

			if droplet.Status == desiredState {
				result <- nil
				return
			}

			// Wait 3 seconds in between
			time.Sleep(3 * time.Second)

			// Verify we shouldn't exit
			select {
			case <-done:
				// We finished, so just exit the goroutine
				return
			default:
				// Keep going
			}
		}
	}()

	log.Printf("Waiting for up to %d seconds for droplet to become %s", timeout/time.Second, desiredState)
	select {
	case err := <-result:
		return err
	case <-time.After(timeout):
		err := fmt.Errorf("Timeout while waiting to for droplet to become '%s'", desiredState)
		return err
	}
}

// waitForActionState simply blocks until the droplet action is in
// a state we expect, while eventually timing out.
func waitForActionState(
	desiredState string, dropletId, actionId int,
	client *godo.Client, timeout time.Duration) error {
	done := make(chan struct{})
	defer close(done)

	result := make(chan error, 1)
	go func() {
		attempts := 0
		for {
			attempts += 1

			log.Printf("Checking action status... (attempt: %d)", attempts)
			action, _, err := client.DropletActions.Get(context.TODO(), dropletId, actionId)
			if err != nil {
				result <- err
				return
			}

			if action.Status == desiredState {
				result <- nil
				return
			}

			// Wait 3 seconds in between
			time.Sleep(3 * time.Second)

			// Verify we shouldn't exit
			select {
			case <-done:
				// We finished, so just exit the goroutine
				return
			default:
				// Keep going
			}
		}
	}()

	log.Printf("Waiting for up to %d seconds for action to become %s", timeout/time.Second, desiredState)
	select {
	case err := <-result:
		return err
	case <-time.After(timeout):
		err := fmt.Errorf("Timeout while waiting to for action to become '%s'", desiredState)
		return err
	}
}

// WaitForImageState simply blocks until the image action is in
// a state we expect, while eventually timing out.
func WaitForImageState(
	desiredState string, imageId, actionId int,
	client *godo.Client, timeout time.Duration) error {
	done := make(chan struct{})
	defer close(done)

	result := make(chan error, 1)
	go func() {
		attempts := 0
		for {
			attempts += 1

			log.Printf("Checking action status... (attempt: %d)", attempts)
			action, _, err := client.ImageActions.Get(context.TODO(), imageId, actionId)
			if err != nil {
				result <- err
				return
			}

			if action.Status == desiredState {
				result <- nil
				return
			}

			// Wait 3 seconds in between
			time.Sleep(3 * time.Second)

			// Verify we shouldn't exit
			select {
			case <-done:
				// We finished, so just exit the goroutine
				return
			default:
				// Keep going
			}
		}
	}()

	log.Printf("Waiting for up to %d seconds for image transfer to become %s", timeout/time.Second, desiredState)
	select {
	case err := <-result:
		return err
	case <-time.After(timeout):
		err := fmt.Errorf("Timeout while waiting to for image transfer to become '%s'", desiredState)
		return err
	}
}
