package main

import(
	"fmt"
	dem "github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs/common"
	"github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs/events"
	"math"
	"os"
	"strconv"
)

const samplesPerSecond = 32
const secondsBeforeAttack = 5
const secondsAfterAttack = 1
const secondsPerAttack = secondsBeforeAttack + secondsAfterAttack
const samplesPerAttack = int(samplesPerSecond * secondsPerAttack)

type PlayerData struct {
	ViewDirectionY float32
	ViewDirectionX float32
	ammoLeft int
	nSpotted int
	score int
	ping int
}

type AttackTime struct {
	attacker    int
	victim      int
	startFrame  int
	attackFrame int
	endFrame    int
	aSpottedV bool
	vSpottedA bool
}

type FireFrameKey struct {
	shooter int
	frame   int
}

var validGuns = map[string]bool{
	"AK-47": true,
	"M4A4": true,
	"AWP": true,
}

func parseAttackData(filename string) {
	var suspect = uint64(42)

	var attackTimes []AttackTime{}

	var fireFrames = map[FireFrameKey]bool{}
	var isMarked = map[int]bool{}

	var markedFrameData = map[int]map[int]PlayerData{}

	f, err := os.Open(filename)
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)
	if err != nil {
		panic(err)
	}

	p := dem.NewParser(f)

	h, err := p.ParseHeader()
	FrameRate := h.FrameRate()

	// Calculate the demo framerate with some hacks
	tick := -1
	for !(2900 < tick && tick < 3000) {
		_, err = p.ParseNextFrame()
		tick = p.GameState().IngameTick()
	}
	if err != nil {
		panic(err)
	}

	iters := 10
	for i:=0; i<iters; i++ {
		_, err = p.ParseNextFrame()
		if err != nil {
			panic(err)
		}
	}
	nextTick := p.GameState().IngameTick()

	TicksPerFrame := float64(nextTick - tick) / float64(iters)

	FrameRate2 := p.TickRate() / TicksPerFrame

	if FrameRate == 0 {
		FrameRate = FrameRate2
	}

	var framesBeforeAttack int
	var framesAfterAttack int
	if (math.Abs(FrameRate-32.0) < 1) && (FrameRate2 == 32) {
		framesBeforeAttack = secondsBeforeAttack * 32
		framesAfterAttack = secondsAfterAttack * 32
	} else if (math.Abs(FrameRate - 64.0) < 4) && (FrameRate2 == 64) {
		framesBeforeAttack = secondsBeforeAttack * 64
		framesAfterAttack = secondsAfterAttack * 64
	} else if (math.Abs(FrameRate - 128) < 4) && (FrameRate2 == 128) {
		framesBeforeAttack = secondsBeforeAttack * 128
		framesAfterAttack = secondsAfterAttack * 128
	} else {
		fmt.Println("Invalid frame rate: ", FrameRate, FrameRate2)
		return
	}

	framesPerAttack := framesBeforeAttack + framesAfterAttack
	framesPerSample := int(framesPerAttack / samplesPerAttack)
	fmt.Println("Frames per sample ", framesPerSample)


	attackCount := 0
	var start int = 0
	var end int = 0
	var frame int = 0
	var attackFrame int
	p.RegisterEventHandler(func(e events.PlayerHurt) {
		if !validGuns[e.Weapon.String()] {
			return
		}
		if e.Attacker.SteamID64 == 0 {
			return
		}

		attackCount++
		attackFrame := p.CurrentFrame()
		start := attackFrame - framesBeforeAttack
		end := attackFrame + framesAfterAttack
		for frame := start; frame < end; frame++ {
			isMarked[frame] = true
		}
		isMarked[start-framesPerSample] = true		// For first sample data angles
		aSpottedV := e.Attacker.HasSpotted(e.Player)
		vSpottedA := e.Player.HasSpotted(e.Attacker)
		new := AttackTime{
			e.Attacker.UserID, e.Player.UserID, start, attackFrame, end, aSpottedV, vSpottedA,
		}
		attackTimes = append(attackTimes, new)
	})

	var i int = 0
	p.RegisterEventHandler(func(e events.WeaponFire) {
		frame = p.CurrentFrame()
		for i=0; i<framesPerSample; i++ {
			fireFrames[FireFrameKey{e.Shooter.UserID, frame - i}] = true
		}
	})

	err = p.ParseToEnd()

	f, err = os.Open(filename)
	p = dem.NewParser(f)
	var ok bool
	for ok=true; ok; ok, err = p.ParseNextFrame() {
		if err != nil {
			panic(err)
		}
		frame = p.CurrentFrame()
		if !isMarked[frame] {
			continue
		}
		var players = map[int]PlayerData{}
		gs := p.GameState()

		for _, player := range gs.Participants().Playing() {
			hasSpotted := gs.Participants().SpottedBy(player)
			players[player.UserID] = extractPlayersData(hasSpotted, frame, player, fireFrames)
		}
	}
}
func extractPlayersData(
	spotted []*common.Player,
	frame int,
	player *common.Player,
	fireFrames map[FireFrameKey]bool) PlayerData {
	return []string{
		strconv.Itoa(frame),
		player.Name,
		strconv.FormatUint(player.SteamID64, 10),
		strconv.FormatFloat(player.Position().X, 'G', -1, 64),
		strconv.FormatFloat(player.Position().Y, 'G', -1, 64),
		strconv.FormatFloat(player.Position().Y, 'G', -1, 64),

		strconv.FormatFloat(player.LastAlivePosition.X, 'G', -1, 64),
		strconv.FormatFloat(player.LastAlivePosition.Y, 'G', -1, 64),
		strconv.FormatFloat(player.LastAlivePosition.Z, 'G', -1, 64),

		strconv.FormatFloat(player.Velocity().X, 'G', -1, 64),
		strconv.FormatFloat(player.Velocity().Y, 'G', -1, 64),
		strconv.FormatFloat(player.Velocity().Z, 'G', -1, 64),

		strconv.FormatFloat(float64(player.ViewDirectionX()), 'G', -1, 64),
		strconv.FormatFloat(float64(player.ViewDirectionY()), 'G', -1, 64),

		strconv.Itoa(player.Health()),
		strconv.Itoa(player.Armor()),
		strconv.Itoa(player.Money()),
		strconv.Itoa(player.EquipmentValueCurrent()),
		strconv.Itoa(player.EquipmentValueFreezeTimeEnd()),
		strconv.Itoa(player.EquipmentValueRoundStart()),
		strconv.FormatBool(player.IsDucking()),
		strconv.FormatBool(player.HasDefuseKit()),
		strconv.FormatBool(player.HasHelmet()),
		strconv.Itoa(player.Kills()),
		strconv.Itoa(player.Deaths()),
		strconv.Itoa(player.Assists()),
		strconv.Itoa(player.Score()),
		strconv.Itoa(player.MVPs()),
		strconv.Itoa(player.MoneySpentTotal()),
		strconv.Itoa(player.MoneySpentThisRound()),
	}
}
