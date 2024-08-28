import { ReconnectingWebsocket } from "./reconnecting_websocket.js";


/**
 * Fetchs session from url
 * @returns {string}
 */
function getSessionId(){
    const path = location.pathname.split("/").filter(Boolean);
    return path.at(-1)
}


class Client {


    /**@type {ReconnectingWebsocket} */
    #ws
    /**@type {string} */
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
    }
    /**
     * 
     * @param {WebsocketMessageEvent} msg 
     */
    onSocketMessage = (msg) => {
        try {
            console.log(msg);
            const data = JSON.parse(msg.detail.data);
            switch (data.contentType) {
                case "BoradcastMessage": {
                    const feed = document.getElementById("global-feed");

                    const el = document.createElement("div");
                    el.style = "margin: 5px 8px; border: black solid 1px;padding: 2px 4px;"

                    el.addEventListener("click",()=>{

                        const dialog = showModal("message-content")
                        dialog.querySelector("p").textContent = data.content;
                    });


                    const header = document.createElement("h4");
                    header.style = "margin-bottom:0;margin-top:2px;"
                    header.textContent = "Notification";

                    el.appendChild(header);

                    el.appendChild(document.createElement("hr"));

                    const text = document.createElement("p");
                    text.style = "margin-top:0px;-webkit-line-clamp:2;overflow:hidden;-webkit-box-orient:vertical;display:-webkit-box;";
                    text.textContent = data.content;

                    el.appendChild(text);

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
     * @param {WebsocketErrorEvent} err 
     */
    onSocketError = (err) => console.error(err);
    init(){
        this.#ws = new ReconnectingWebsocket(`ws://${document.location.host}/session/${this.#session}/ws`);
        this.#ws.addEventListener("close",this.onSocketClose);
        this.#ws.addEventListener("error",this.onSocketError);
        this.#ws.addEventListener("message",this.onSocketMessage);
        this.#ws.addEventListener("open",this.onSocketOpen);
        this.#ws.addEventListener("reconnecting",(ev)=>console.log(ev));
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

    /**@type {HTMLFormElement} */
    const form = ev.target
    form.reset();

    const msg = data.get("msg");
    client.broadcast(msg);
}

/**
 * 
 * @param {string} id 
 */
function showModal(id){
    /**@type {HTMLDialogElement} */
    const dialog = htmx.find(`#${id}`)
    dialog.showModal();
    return dialog;
}

window.traveller = {
    broadcastMessage,
    showModal
}