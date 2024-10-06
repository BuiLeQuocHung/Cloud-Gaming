import useWebSocket from 'react-use-websocket';
import { WebSocketHook } from 'react-use-websocket/dist/lib/types';

let connection: WebSocketHook

const useWebSocketConnection = (url) => {
  connection = useWebSocket(url, {
    shouldReconnect: (closeEvent) => {
      return closeEvent.code !== 1000
    },
    onOpen: () => console.log('WebSocket connection established'),
    onClose: () => console.log('WebSocket connection closed'),
    onError: (error) => console.error('WebSocket error:', error),
  });

  return connection
};

export const websocket = {
  init: useWebSocketConnection,
  connection: () => connection,
};
