package steps

import (
	"artifactsmmo.com/m/api/actions/npc"
	"artifactsmmo.com/m/game"
)

func NPCBuy(kernel *game.Kernel, code string, quantity int) error {
	log := kernel.LogPreF("[%s]<npc-buy>", kernel.CharacterName)

	log("buying %d %s", quantity, code)
	res, err := npc.NPCBuyItem(kernel.CharacterName, code, quantity)
	if err != nil {
		log("failed to buy %d %s from npc", quantity, code)
		return err
	}

	kernel.CharacterState.Set(&res.Character)
	kernel.WaitForDown(res.Cooldown)

	return nil
}

func NPCSell(kernel *game.Kernel, code string, quantity int) error {
	log := kernel.LogPreF("[%s]<npc-sell>", kernel.CharacterName)

	log("selling %d %s", quantity, code)
	res, err := npc.NPCSellItem(kernel.CharacterName, code, quantity)
	if err != nil {
		log("failed to sell %d %s to npc", quantity, code)
		return err
	}

	kernel.CharacterState.Set(&res.Character)
	kernel.WaitForDown(res.Cooldown)

	return nil
}
