import { WebSocketHook } from "react-use-websocket/dist/lib/types";
import { ResponseMessage, RequestMessage, MSG_COOR_HANDSHAKE, MSG_WEBRTC_ANSWER, MSG_WEBRTC_ICE_CANDIDATE, MSG_WEBRTC_OFFER } from "../message/message.ts";
import { useEffect } from "react";
import eventBus from "../bus/bus.ts";
import { websocket } from "../network/websocket.ts";
import { webrtc } from "../network/webrtc.ts";
import { base64ToString } from "../util/decoder.ts";
import { stringToBase64 } from "../util/encoder.ts";
import { appWebRTC } from "./webrtc.ts";

export const ListenWebsocket = () => {
  const signalingServer: WebSocketHook = websocket.connection()
  const peerConnection: RTCPeerConnection = webrtc.connection()

  useEffect(() => {
      const handleMessage = async () => {
        let message = signalingServer.lastMessage;

        if (message === null) {
          return;
        }
  
        await handleSignalingMessage(message, peerConnection, signalingServer);
      };
  
      handleMessage();
      
    }, [signalingServer, signalingServer.lastMessage, peerConnection, ]);
  };
  
const handleSignalingMessage = async (message: MessageEvent<any>, peerConnection: RTCPeerConnection, signalingServer: WebSocketHook) => {
    console.log(message)
    let data: ResponseMessage = JSON.parse(message.data) as ResponseMessage
    data.payload = base64ToString(data.payload)
    console.log("label: ", data.label, data.payload)

    switch (data.label) {
        case MSG_COOR_HANDSHAKE:
            eventBus.emit('gameList', data.payload)
            break;
        case MSG_WEBRTC_OFFER:
            const offer: RTCSessionDescriptionInit = JSON.parse(data.payload) as RTCSessionDescriptionInit
            console.log("offer: ", offer)

            await peerConnection.setRemoteDescription(offer);

            const answer = await peerConnection.createAnswer()
            await peerConnection.setLocalDescription(answer);

            let new_message: RequestMessage = {
                label: MSG_WEBRTC_ANSWER,
                payload: stringToBase64(JSON.stringify(answer)),
            }

            signalingServer.sendJsonMessage(new_message)

            appWebRTC.updateFlagRemoteDescription(true)

            break;

        case MSG_WEBRTC_ICE_CANDIDATE:
            const candidate: RTCIceCandidateInit = JSON.parse(data.payload) as RTCIceCandidateInit
            // Do NOT add candidate directly to peer
            // The addIceCandidate() method is only valid after the remote description has been set.
            // InvalidStateError: Failed to execute 'addIceCandidate' on 'RTCPeerConnection': The remote description was null
            appWebRTC.addIceCandidate(candidate)
            break;
    }
}