import React, { useContext, useEffect, useRef } from 'react';
import { keyboardPressed, keyboardReleased } from './keyboard.ts';
import { webrtc } from '../network/webrtc.ts';
import { ParentContext } from '../App/App.js';


const Streaming = ({mediaStream}) => {
  const { setState } = useContext(ParentContext);
  const videoRef = useRef(null);
  console.log("media stream: ", mediaStream)

  useEffect(() => {
    if (mediaStream && videoRef.current) {
      videoRef.current.srcObject = mediaStream;
      videoRef.current.volumn = 1
      console.log("media tracks: ", mediaStream.getTracks());
      videoRef.current.play().catch((error) => {
        console.error('Error starting video playback:', error);
      })
    }
  }, [mediaStream]);

  useEffect(()=>{
    document.getElementById('media').focus()
  })

  const peerConnection = webrtc.connection()
    setInterval(() => {
      peerConnection.getStats().then(stats => {
          stats.forEach(report => {
              if (report.type === "inbound-rtp" && report.kind === "video") {
                  console.log("Video Stats:", report);
              }
              if (report.type === "inbound-rtp" && report.kind === "audio") {
                console.log("Audio Stats:", report);
            }
          });
        });
      console.log(mediaStream.getVideoTracks()[0].getSettings().frameRate)
  }, 1000);

  

  return (
    <div id='media'
      tabIndex={0}
      onKeyDown={(event) => keyboardPressed(event, setState)}
      onKeyUp={(event) => keyboardReleased(event, setState)}
      style={{width:"auto", height:"auto"}}
    >
      <video ref={videoRef} autoPlay playsInline></video>
    </div>
  );
};

export default Streaming

