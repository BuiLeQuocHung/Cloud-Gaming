import { WebSocketHook } from "react-use-websocket/dist/lib/types";
import { RequestMessage, MSG_WEBRTC_INIT } from "../message/message.ts";
import { websocket } from "../network/websocket.ts";
import { stringToBase64 } from "../util/encoder.ts";
import { webrtc } from "../network/webrtc.ts";

let iceCandidateQueue: Array<RTCIceCandidateInit> = []
let isRemoteDescriptionSet: boolean = false

const webRTCHandshake = () => {
    const signalingServer: WebSocketHook = websocket.connection()
    let initMessage: RequestMessage = {
        label: MSG_WEBRTC_INIT,
        payload: stringToBase64(JSON.stringify({})),
    }

    signalingServer.sendJsonMessage(initMessage)
}

const addIceCandidate = (candidate: RTCIceCandidateInit) => {
    if (isRemoteDescriptionSet) {
        const pc: RTCPeerConnection = webrtc.connection()
        pc.addIceCandidate(candidate)
    } else {
        iceCandidateQueue.push(candidate)
    }
}

const addIceCandidateFromQueueToPeerConnection = async () => {
    const pc: RTCPeerConnection = webrtc.connection()
    console.log("addIceCandidateFromQueueToPeerConnection: ", iceCandidateQueue)
    await Promise.all(iceCandidateQueue.map((candidate) => pc.addIceCandidate(candidate)))
    iceCandidateQueue = [] // reset queue
}

const updateFlagRemoteDescription = (isSet: boolean) => {
    isRemoteDescriptionSet = isSet
    if (isRemoteDescriptionSet) {
        addIceCandidateFromQueueToPeerConnection()
    }
}


export const appWebRTC = {
    handshake: webRTCHandshake,
    addIceCandidate: addIceCandidate,
    updateFlagRemoteDescription: updateFlagRemoteDescription,
}