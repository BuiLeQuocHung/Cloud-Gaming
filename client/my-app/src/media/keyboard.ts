
import { webrtc } from '../network/webrtc.ts';
import { map } from '../input/input.ts';
import { ButtonState, KeyboardData, MSG_STOP_GAME, RequestMessage } from '../message/message.ts';
import { websocket } from '../network/websocket.ts';
import { WebSocketHook } from 'react-use-websocket/dist/lib/types';
import { DISPLAY_MENU } from '../App/state.ts';

export const keyboardPressed = (keyboardEvent, setState) => {
    console.log("keyboard pressed: ", keyboardEvent.code)
    if (keyboardEvent.code === "KeyQ") {
        const signalingServer: WebSocketHook = websocket.connection()
        const msg: RequestMessage = {
            label: MSG_STOP_GAME,
            payload: "",
        }

        signalingServer.sendMessage(JSON.stringify(msg))
        setState(DISPLAY_MENU)
        return
    }

    const btnId: number = map[keyboardEvent.code]
    const buttonState: ButtonState = {
        button: btnId,
        pressed: true,
    }

    const data: KeyboardData = {
        user: 0, // currently only support 1 player
        button_state: [buttonState],
    }

    webrtc.inputChannel.sendKeyBoardInput(JSON.stringify(data))
}

export const keyboardReleased = (keyboardEvent, setState) => {
    console.log("keyboard released: ", keyboardEvent.code)
    if (keyboardEvent.code === "KeyQ") {
        return
    }
    const btnId: number = map[keyboardEvent.code]
    const buttonState: ButtonState = {
        button: btnId,
        pressed: false,
    }

    const data: KeyboardData = {
        user: 0, // currently only support 1 player
        button_state: [buttonState],
    }

    console.log("data: ", data)
    webrtc.inputChannel.sendKeyBoardInput(JSON.stringify(data))
}