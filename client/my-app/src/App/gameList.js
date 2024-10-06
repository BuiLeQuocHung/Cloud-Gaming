import React, { useContext, useEffect, useState } from "react"
import eventBus from "../bus/bus.ts"
import { websocket } from "../network/websocket.ts"
import { MSG_START_GAME } from "../message/message.ts"
import { stringToBase64 } from "../util/encoder.ts"
import { ParentContext } from "./App.js"
import { DISPLAY_STREAMING } from "./state.ts"
import FocusLock from 'react-focus-lock'

const GameList = () => {
    const { setState } = useContext(ParentContext);
    
    const [gameList, setGameList] = useState([])
    const [itemSelect, setItemSelect] = useState(0)

    useEffect(() => {
        const loadData = sessionStorage.getItem('gameList')
        if (loadData) {
            setGameList(JSON.parse(loadData))
        } else {
            eventBus.once('gameList', (data) => {
                setGameList(JSON.parse(data))
                sessionStorage.setItem('gameList', data)
            })}
    }, [])

    useEffect(()=> {
        document.getElementById("game-list").focus()
    })

    

    const handleItemSelect = (keyboardEvent) => {
        let n = gameList.length
        if (keyboardEvent.code === 'ArrowUp') {
            setItemSelect(itemSelect-1 >= 0 ? itemSelect-1 : n-1)
        } else if (keyboardEvent.code === 'ArrowDown') {
            setItemSelect((itemSelect+1)%n)
        } else if (keyboardEvent.code === "Enter") {
            handleItemClick({item: gameList[itemSelect], setState: setState})
        }
    }

    // document.getElementById("game-list").addEventListener("keydown", handleItemSelect)
    
    return (
        <FocusLock>
        <div id="game-list"
            tabIndex={0}
            onKeyDown={handleItemSelect}
            style={{border:"none"}}
        >
            <ul>
                {gameList.map((item, index) => (
                    <li style={{color: index === itemSelect ? "red" : "blue"}}  key={index} onClick={() => handleItemClick({item, setState})}>{item.name}</li>
                ))}
            </ul>
        </div>
        </FocusLock>
    )
}

const handleItemClick = ({item, setState}) => {
    console.log(item)
    const signalingServer = websocket.connection()
    signalingServer.sendJsonMessage({
        label: MSG_START_GAME,
        payload: stringToBase64(JSON.stringify({
            game: item.name,
        })),
    })
    setState(DISPLAY_STREAMING)
}





export default GameList