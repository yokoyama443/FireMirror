package models

import (
	"log"
	"time"
)

type AttackList struct {
	ID       int
	CreateAt string
	UpdateAt string
	IP       string
	Status   string //Waiting, Attacking, Success, Failed
}

func (a *AttackList) Insert() error {
	cmd := `INSERT INTO attack_list (CreateAt, UpdateAt, IP, Status) VALUES (?, ?, ?, ?)`
	a.CreateAt = time.Now().Format("2006-01-02 15:04:05")
	a.UpdateAt = ""
	_, err := DB.Exec(cmd, a.CreateAt, a.UpdateAt, a.IP, a.Status)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	return nil
}

func (a *AttackList) Update() error {
	cmd := `UPDATE attack_list SET Status = ? WHERE IP = ?`
	a.UpdateAt = time.Now().Format("2006-01-02 15:04:05")
	_, err := DB.Exec(cmd, a.Status, a.IP)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	return nil
}

func GetAllAttackList() (attackLists []AttackList, err error) {
	cmd := `SELECT * FROM attack_list`
	rows, err := DB.Query(cmd)
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var attackList AttackList
		err = rows.Scan(&attackList.ID, &attackList.CreateAt, &attackList.UpdateAt, &attackList.IP, &attackList.Status)
		if err != nil {
			log.Fatalln(err)
			return
		}
		attackLists = append(attackLists, attackList)
	}
	return
}

func GetWaitingAttackList() (AttackList, error) {
	cmd := `SELECT * FROM attack_list WHERE Status = ? LIMIT 1`
	rows, err := DB.Query(cmd, "Waiting")
	if err != nil {
		log.Fatalln(err)
		return AttackList{}, err
	}
	defer rows.Close()
	var attackList AttackList
	for rows.Next() {
		err = rows.Scan(&attackList.ID, &attackList.CreateAt, &attackList.UpdateAt, &attackList.IP, &attackList.Status)
		if err != nil {
			log.Fatalln(err)
			return AttackList{}, err
		}
	}
	return attackList, nil
}
