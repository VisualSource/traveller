"use strict"

/**
 * Fetchs session from url
 * @returns {string}
 */
function getSessionId(){
    const path = location.pathname.split("/").filter(Boolean);
    return path.at(-1)
}


class Client {
    /** @type {Websocket} */
    #ws
    /** @type {string} */
    #session
    constructor(){
        /**
         * @type {string}
         */
        this.#session = getSessionId();
    }

    onSocketOpen = () => {
        console.info("Websocket has opened!");
    }
    onSocketClose = (ev) => {
        console.info("Websocket has closed!");
        setTimeout(()=>{
            this.init();
        },5000);
    }
    /**
     * 
     * @param {} msg 
     */
    onSocketMessage = (msg) => {
        try {
            console.log(msg);
            const data = JSON.parse(msg.data);
            switch (data.contentType) {
                case "BoradcastMessage": {
                    const feed = document.getElementById("global-feed");

                    const el = document.createElement("div");
                    el.textContent = data.content;
                    feed.appendChild(el);
                    break;
                }
                case "PrivateMessage":{
                    break;
                }
                default:
                    break;
            }
        } catch (error) {
            console.error(error);
        }
    }
    /**
     * 
     * @param {unknown} err 
     */
    onSocketError = (err) => {
        console.error(err);
        this.#ws.close();
    }
    init(){
        this.#ws = new WebSocket(`ws://${document.location.host}/session/${this.#session}/ws`);
        this.#ws.addEventListener("close",this.onSocketClose);
        this.#ws.addEventListener("error",this.onSocketError);
        this.#ws.addEventListener("message",this.onSocketMessage);
        this.#ws.addEventListener("open",this.onSocketOpen);
    }
    /**
     * Send a message to all in session
     * @param {string} message
     */
    broadcast(message){
        if(!this.#ws) throw new Error("Websocket is not ready!");
        this.#ws.send(JSON.stringify({ contentType: "BroadcastMessage", payload: { Message: message, Target: "NONE" } }));
    }
    /**
     * Send a message to a single target
     * @param {string} message
     * @param {string} userId
     * @returns {void} 
     */
    message(userId, message){
        if(!this.#ws) throw new Error("Websocket is not ready!");
        this.#ws.send(JSON.stringify({ contentType:"PrivateMessage",  payload: { Message: message, Target: userId } }));
    }
}

const client = new Client();
client.init();



/**
 * 
 * @param {SubmitEvent} ev 
 */
function broadcastMessage(ev){
    const data = new FormData(ev.target);
    const msg = data.get("msg");
    client.broadcast(msg);
}