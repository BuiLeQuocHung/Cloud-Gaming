# About
This is my personal project about cloud gaming

# Credit
Shout out to 
- https://github.com/giongto35/cloud-game, which helps me to understand the architecture of cloud gaming
- https://github.com/libretro/ludo, a great repo to understand libretro
- https://github.com/pion/webrtc, a great library for webrtc connection

# Run
- Only work on Linux/Window
- If you haven't installed docker yet, please install it at https://www.docker.com/
### From root directory:

**Server**
```
docker-compose -f ./docker/compose.yaml up --build
```

**Client**

```
cd client/my-app/
npm install
npm start
```
