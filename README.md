collabpaint
===========

A collaborative paint web application that leverages Go, HTML5 Canvas,
and WebSockets.

Example usage:

    export GOPATH=$HOME/go
    go get github.com/gorilla/websocket
    go run paintserver.go

... then navigate to http://127.0.0.1:8080/. To see the collaborative
powers, open http://127.0.0.1:8080/ in two or more browser
windows. Changes made in one window should become visible in the other
windows.

This is still in a very nascant state. As time goes on, I hope to make
it better.
