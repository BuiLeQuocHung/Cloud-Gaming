
type MsgType = string

export const MSG_COOR_HANDSHAKE : MsgType = "msg_coor_handshake"

export const MSG_WEBRTC_INIT            : MsgType = "msg_webrtc_init"
export const MSG_WEBRTC_OFFER           : MsgType = "msg_webrtc_offer"
export const MSG_WEBRTC_ANSWER          : MsgType = "msg_webrtc_answer"
export const MSG_WEBRTC_ICE_CANDIDATE   : MsgType = "msg_webrtc_ice_candidate"

export const MSG_START_GAME     : MsgType = "msg_start_game"
export const MSG_STOP_GAME      : MsgType = "msg_stop_game"

export type ResponseMessage = {
    label: MsgType
    payload: string
    error: string
}

export type RequestMessage = {
    label: MsgType
    payload: string
}


export type KeyboardData = {
    user: number
    button_state: ButtonState[]
}

export type ButtonState = {
    button: number
    pressed: boolean
}