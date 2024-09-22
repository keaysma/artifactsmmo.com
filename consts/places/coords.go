package coords

type Coord struct {
	X    int
	Y    int
	Name string
}

// Places
var Spawn = Coord{X: 0, Y: 0, Name: "Spawn"}
var Bank = Coord{X: 4, Y: 1, Name: "Bank"}

// // Fight
var Chickens = Coord{X: 0, Y: 1, Name: "Chickens"}
var Yellow_Slimes = Coord{X: 4, Y: -1, Name: "Yellow Slimes"}

// // Gather
var AshTree = Coord{X: -1, Y: 0, Name: "Ash Tree"}
var CopperRocks = Coord{X: 2, Y: 0, Name: "Copper Rocks"}
var GudgeonFishingSpot = Coord{X: 4, Y: 2, Name: "Gudgeon Fishing Spot"}
var ShrimpFishingSpot = Coord{X: 5, Y: 2, Name: "Shrimp Fishing Spot"}

// // Money
var Bank_City = Coord{X: 4, Y: 1, Name: "Bank in City"}
var GrandExchange = Coord{X: 5, Y: 1, Name: "Grand Exchange"}

// // Skills
var WeaponCrafting_City = Coord{X: 2, Y: 1, Name: "Weapon Crafting in City"}
var MiningWorkshop = Coord{X: 1, Y: 5, Name: "Mining Workshop"}
