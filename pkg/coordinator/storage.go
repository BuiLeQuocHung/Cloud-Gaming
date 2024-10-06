package coordinator

type (
	GameInfo struct {
		Name string `json:"name"`
	}
)

func (c *Coordinator) getListGames() []GameInfo {
	allGamesMetadata := c.storage.GetAllGamesMetadata()
	listGames := make([]GameInfo, 0, len(allGamesMetadata))

	for _, gameMeta := range allGamesMetadata {
		listGames = append(listGames, GameInfo{
			Name: gameMeta.Name,
		})
	}

	return listGames
}
