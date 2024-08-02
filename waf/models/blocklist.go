package models

import (
	"log"
	"time"
)

type BlockList struct {
	ID        int
	CreateAt  string
	UpdateAt  string
	IP        string
	Count     int
	LastEvent string
}

func (b *BlockList) Insert() error {
	cmd := `INSERT INTO block_list (CreateAt, UpdateAt, IP, Count, LastEvent) VALUES (?, ?, ?, ?, ?)`
	now := time.Now().Format("2006-01-02 15:04:05")
	b.CreateAt = now
	b.UpdateAt = now
	b.Count = 1
	b.LastEvent = now
	_, err := DB.Exec(cmd, b.CreateAt, b.UpdateAt, b.IP, b.Count, b.LastEvent)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (b *BlockList) Update() error {
	cmd := `UPDATE block_list SET Count = ?, LastEvent = ?, UpdateAt = ? WHERE IP = ?`
	now := time.Now().Format("2006-01-02 15:04:05")
	b.UpdateAt = now
	b.LastEvent = now
	b.Count++
	_, err := DB.Exec(cmd, b.Count, b.LastEvent, b.UpdateAt, b.IP)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func GetBlockListByIP(ip string) (BlockList, error) {
	cmd := `SELECT * FROM block_list WHERE IP = ?`
	row := DB.QueryRow(cmd, ip)
	var b BlockList
	err := row.Scan(&b.ID, &b.CreateAt, &b.UpdateAt, &b.IP, &b.Count, &b.LastEvent)
	if err != nil {
		return BlockList{}, err
	}
	return b, nil
}

func GetRecentBlockListByIP(ip string, duration time.Duration) (BlockList, error) {
	cmd := `SELECT * FROM block_list WHERE IP = ? AND LastEvent > ? ORDER BY LastEvent DESC LIMIT 1`
	threshold := time.Now().Add(-duration).Format("2006-01-02 15:04:05")
	row := DB.QueryRow(cmd, ip, threshold)
	var b BlockList
	err := row.Scan(&b.ID, &b.CreateAt, &b.UpdateAt, &b.IP, &b.Count, &b.LastEvent)
	if err != nil {
		return BlockList{}, err
	}
	return b, nil
}
