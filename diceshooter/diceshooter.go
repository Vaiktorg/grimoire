package main

//
//import (
//	"encoding/json"
//	"fmt"
//	"os"
//
//	"github.com/vaiktorg/grimoire/helpers"
//
//	"github.com/Pallinder/go-randomdata"
//	"github.com/vaiktorg/grimoire/dice"
//)
//
//type (
//	Thug struct {
//		Type    string         `json:"name"`
//		Plays   map[string]int `json:"record"`
//		Money   int            `json:"money"`
//		Bet     int            `json:"bet"`
//		CanPlay bool           `json:"can_play"`
//	}
//	DiceShooting struct {
//		dice          dice.Dice
//		Players       map[string]*Thug `json:"players"`
//		owner         *Thug
//		opponent      *Thug
//		TotalGames    int           `json:"total_games"`
//		HistoricGames []GameHistory `json:"historic_games"`
//		prevs         []*Thug
//	}
//
//	GameHistory struct {
//		Owner    string `json:"owner"`
//		Opponent string `json:"opponent"`
//		Status   string `json:"status"`
//		Roll     int    `json:"roll"`
//	}
//)
//
//const (
//	Lose = "lose"
//	Win  = "win"
//)
//
//func main() {
//	g := DiceShooting{
//		dice: dice.Dice{},
//	}
//	g.Init(10, 50, 1)
//	g.Play()
//
//	j, err := json.MarshalIndent(g, "", "	")
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	file, err := os.Create("diceshooter_" + helpers.MakeTimestampNum() + ".json")
//	if err != nil {
//		fmt.Println(err)
//		os.Exit(1)
//	}
//	defer file.Close()
//	_, err = file.Write(j)
//	if err != nil {
//		fmt.Println(err)
//		os.Exit(1)
//	}
//
//}
//
//func (g *DiceShooting) Init(playerNum, initialBank, bet int) {
//	g.Players = make(map[string]*Thug)
//	for i := 0; i < playerNum; i++ {
//		thug := &Thug{
//			Type:    randomdata.SillyName(),
//			Plays:   make(map[string]int),
//			Money:   initialBank,
//			Bet:     bet,
//			CanPlay: true,
//		}
//		g.Players[thug.Type] = thug
//	}
//
//	g.owner = g.nextThug()
//	g.opponent = g.nextThug()
//}
//
//func (g *DiceShooting) Play() {
//	for len(g.Players) <= 1 {
//		if g.owner == nil || g.opponent == nil {
//			return
//		}
//
//		if g.owner.Money <= 0 {
//			g.owner.Bet = 0
//			g.owner.CanPlay = false
//			delete(g.Players, g.owner.Type)
//
//			g.owner = g.nextThug()
//		}
//
//		if g.opponent.Money <= 0 {
//			g.opponent.Bet = 0
//			g.opponent.CanPlay = false
//			delete(g.Players, g.opponent.Type)
//
//			g.opponent = g.nextThug()
//		}
//
//		var HGame GameHistory
//		HGame.Opponent = g.opponent.Type
//		HGame.Owner = g.owner.Type
//
//		rollDice := func() int { return g.dice.D6() + g.dice.D6() }
//		roll := rollDice()
//
//		switch roll {
//		case 7, 11:
//			g.owner.win(g.opponent.loss())
//
//			g.opponent = g.nextThug()
//			HGame.Status = Win
//		case 2, 3, 12:
//			g.opponent.win(g.owner.loss())
//
//			g.owner = g.opponent
//			g.opponent = g.nextThug()
//			HGame.Status = Lose
//		case 4, 5, 6, 8, 9, 10:
//			g.owner.Plays["P"]++
//			for {
//				r := rollDice()
//				if r == 7 {
//					g.opponent.win(g.owner.loss())
//					HGame.Status = Lose
//					break
//				} else if r == roll {
//					g.owner.win(g.opponent.loss())
//					g.owner, g.opponent = g.opponent, g.owner
//					HGame.Status = Win
//					break
//				}
//			}
//		}
//		HGame.Roll = roll
//		g.HistoricGames = append(g.HistoricGames, HGame)
//		g.TotalGames++
//	}
//}
//
//func (t *Thug) win(earning int) {
//	t.Plays[Win]++
//	t.Money += earning
//}
//
//func (t *Thug) loss() int {
//	t.Plays[Lose]++
//
//	if t.Money <= t.Bet {
//		t.Bet = t.Money
//		t.Money = 0
//		return t.Bet
//	}
//
//	t.Money -= t.Bet
//	return t.Bet
//}
//
//func (g *DiceShooting) nextThug() *Thug {
//	for _, t := range g.Players {
//		if len(g.prevs) <= 1 {
//			g.prevs = append(g.prevs, t)
//		}
//		if len(g.Players) <= 2 {
//			return t
//		}
//		if g.prevs[0] == t || g.prevs[1] == t {
//			continue
//		}
//		g.prevs = append(g.prevs[1:], t)
//		return t
//	}
//	return nil
//}
