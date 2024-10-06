import { WebSocketHook } from 'react-use-websocket/dist/lib/types';
import { RequestMessage, MSG_WEBRTC_ICE_CANDIDATE } from '../message/message.ts';
import { websocket } from './websocket.ts';
import { stringToBase64 } from '../util/encoder.ts';

let peerConnection: RTCPeerConnection 
let mediaStream: MediaStream 

let keyboardChannel: RTCDataChannel 
let mouseChannel: RTCDataChannel 

const iceServers = {
    iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
  };

const init = () => {
    const signalingServer: WebSocketHook = websocket.connection()
    peerConnection = new RTCPeerConnection(iceServers)
    mediaStream = new MediaStream()

    peerConnection.ondatachannel = (event) => {
        const channel = event.channel
        switch (channel.label) {
            case "keyboard":
                keyboardChannel = channel
                break;
            case "mouse":
                mouseChannel = channel
                break;
            default:
                break;
        }

        channel.onopen = () => {console.log(`${channel.label} channel has been opened`)}
        channel.onclose = () => {console.log(`${channel.label} channel has been closed`)}
    }

    peerConnection.onicecandidate = (event) => {
        let message: RequestMessage = {
            label: MSG_WEBRTC_ICE_CANDIDATE,
            payload: stringToBase64(JSON.stringify(event.candidate)),
        }
        signalingServer.sendJsonMessage(message)
    }

    peerConnection.ontrack = (event) => {
        console.log("track received: ", event.track.label, event.track)
        mediaStream.addTrack(event.track)
    }

    console.log("webrtc how many times")
    return peerConnection
}

const stop = () => {
    if (peerConnection) {
        peerConnection.close()
    }

    if (mediaStream) {
        mediaStream.getTracks().forEach(track => {
            track.stop()
        });
    }

    if (keyboardChannel) {
        keyboardChannel.close()
    }

    if (mouseChannel) {
        mouseChannel.close()
    }
}

export const webrtc = {
    init: init,
    stop: stop,
    connection: () => peerConnection,
    inputChannel: {
        sendKeyBoardInput: (data: string) => keyboardChannel.send(data),
        sendMouseInput: (data: string) => mouseChannel.send(data),
    },
    mediaStream: () => mediaStream,
}
