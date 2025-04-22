````
https://cloud.google.com/livestream/docs/quickstarts/quickstart-dash#stop-channel-drest

ffmpeg -re -f lavfi -i "testsrc=size=1920x1080 [out0]; sine=frequency=500 [out1]" -acodec aac -vcodec h264 -f flv <RTMP_URI>

https://shaka-player-demo.appspot.com/demo/#audiolang=en;textlang=en;uilang=en;panel=CUSTOM%20CONTENT;build=uncompiled


````