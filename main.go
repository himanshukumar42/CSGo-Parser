package main

import (
	"fmt"
	"log"
	"os"

	dem "github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs/events"
)

func main() {
	f, err := os.Open("tv_demo1.dem")
	if err != nil {
		log.Panic("failed to open demo file: ", err)
	}

	p := dem.NewParser(f)
	defer func(p dem.Parser) {
		err := p.Close()
		if err != nil {
			log.Panic("failed to close parser", err)
		}
	}(p)

	p.RegisterEventHandler(func(start events.MatchStart) {
		fmt.Println("------------------------------------")

		fmt.Println("List of all Players:- ")
		allPlayers := p.GameState().Participants().All()
		fmt.Println("No of Players in the Game:- ", len(allPlayers))
		for _, pl := range allPlayers {
			fmt.Println(pl.Name)
		}
		allParticipants := p.GameState().Participants().Playing()
		fmt.Println("No of Players playing and Participating the Game:- ", len(allParticipants))
		fmt.Println("List of Players Participating and Playing the Game:- ")
		for _, pm := range allParticipants {
			fmt.Println(pm.Name)
		}

		fmt.Println(p.CurrentTime())
		fmt.Println(p.CurrentFrame())
		fmt.Println(p.GameState().TeamTerrorists().Members())
		fmt.Println(p.GameState().TeamCounterTerrorists().Members())
		fmt.Println(p.GameState().TeamTerrorists().ClanName())
		fmt.Println("------------------------------------")
	})

	p.RegisterEventHandler(func(e events.Kill) {
		if e.IsHeadshot {
			fmt.Printf("\n%s player killed %v with Weapon %v by Headshot from a distance of %f\n", e.Killer, e.Victim, e.Weapon, e.Distance)
		} else {
			fmt.Printf("\n%s player killed %v with Weapon %v from a distance of %f\n", e.Killer, e.Victim, e.Weapon, e.Distance)
		}
	})

	err = p.ParseToEnd()
	if err != nil {
		log.Panic("failed to parse demo: ", err)
	}
}
