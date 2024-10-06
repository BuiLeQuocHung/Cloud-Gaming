import './App.css';
import Streaming from '../media/media.js'
import { webrtc } from '../network/webrtc.ts';
import { websocket } from '../network/websocket.ts';
import { ListenWebsocket }  from './websocket.ts';
import { appWebRTC } from './webrtc.ts';
import GameList from './gameList.js';
import { DISPLAY_MENU, DISPLAY_STREAMING } from './state.ts';
import React, { createContext, useEffect, useState } from 'react';

const signalingServerAddress = 'ws://localhost:9090/init/user/ws'

export const ParentContext = createContext();

const App = () => {
  const [state, setState] = useState(DISPLAY_MENU)
  websocket.init(signalingServerAddress)
  useEffect(() => {
    webrtc.init()
    appWebRTC.handshake()
  }, [])

  return (
    <div>
    <ListenWebsocket/>
    <ParentContext.Provider value={{ state, setState }}>
      <div className="App">
        <header className="App-header">
          {displayState(state)}
        </header>
      </div>
    </ParentContext.Provider>
    </div>
  );
}

const displayState = (state) => {
  switch (state) {
    case DISPLAY_MENU:
      return <GameList />
    case DISPLAY_STREAMING:
      return <Streaming mediaStream={webrtc.mediaStream()}/>
    default:
      return <GameList/>
  }
}

export default App;