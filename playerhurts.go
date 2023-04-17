package main

import (
	dem "github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs/common"
	"github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs/events"
	"log"
	"os"
)

func fetchPlayerHurt(filename string) {
	f, err := os.Open(filename)
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)
	if err != nil {
		log.Fatalln(err)
	}

	p := dem.NewParser(f)

	p.RegisterEventHandler(func(e events.PlayerHurt) {

	})

	for ok := true; ok; ok, err = p.ParseNextFrame() {
		if err != nil {
			log.Fatalln(err)
		}

	}
}

func playerHurtEvent(frame int, player *common.Player) *events.PlayerHurt {
	return &events.PlayerHurt{
		Player:            player,
		Attacker:          nil,
		Health:            0,
		Armor:             0,
		Weapon:            nil,
		WeaponString:      "",
		HealthDamage:      0,
		ArmorDamage:       0,
		HealthDamageTaken: 0,
		ArmorDamageTaken:  0,
		HitGroup:          0,
	}
}
