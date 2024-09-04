import { ReconnectingWebsocket } from "./reconnecting_websocket.js";
import { makePlanet, Galaxy } from "./galaxy_editor.js";
import { Vector2 , MathUtils } from "three";
/**
 * Fetchs session from url
 * @returns {string}
 */
function getSessionId(){
    const path = location.pathname.split("/").filter(Boolean);
    return path.at(-1)
}

class Client {
    /** @type {Galaxy} */
    system;
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
    onSocketMessage = async (msg) => {
        try {
            console.log(msg);
            const data = JSON.parse(msg.detail.data);
            switch (data.contentType) {
                case "BoradcastMessage": {
                    const feed = document.getElementById("global-feed");

                    const el = document.createElement("div");
                    el.style = "margin: 5px 8px; border: black solid 1px;padding: 2px 4px;"
                    el.setAttribute("data-type","message");

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
                    const target = htmx.find("#private-message-board");

                    // linegreen for current user, blue for other
                    const textboxColor = "limegreen";

                    const el = document.createElement("div");
                    el.style = `margin: 5px 8px; border: black solid 1px;padding: 2px 4px;display:flex; background-color: ${textboxColor};`

                    /*const i = document.createElement("img");
                    i.style="height:45px;width:45px";

                    const encoder = new TextEncoder();
                    const item = encoder.encode("collin_blosser@yahoo.com")

                    const hash = await crypto.subtle.digest("SHA-256",item).then(e=> Array.from(new Uint8Array(e)) ).then(e=>e.map(b=>b.toString(16).padStart(2,"0")).join(""));

                    i.src = `https://www.gravatar.com/avatar/${hash}`;
                    i.alt="user-icon"

                    el.appendChild(i);*/

                    const text = document.createElement("p");
                    text.style = "margin-top:0px;";
                    text.textContent = data.message;

                    el.appendChild(text);

                    target.appendChild(el);

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


    initWorld(){
        /**
         * @type {HTMLDivElement}
         */
        const container = document.getElementById("gallexy-view");

        this.system = new Galaxy(container);

        this.system.init(true);

        this.system.camera.position.set(0,11,10);
        this.system.camera.rotation.set(-0.90,0,0);
    
        const MOON_RADIUS = 0.27;
        const EARTH_RADIUS = 1;

        const earth = makePlanet({
            radius: EARTH_RADIUS,
            mat: {
                specular: 0x333333,
                shininess: 5,
                map: this.system.textureLoader.load("/static/textures/planets/earth_atmos_2048.jpg"),
                specularMap: this.system.textureLoader.load("/static/textures/planets/earth_specular_2048.jpg"),
                normalMap: this.system.textureLoader.load("/static/textures/planets/earth_normal_2048.jpg"),
                normalScale: new Vector2(0.85,0.85)
            },
            label: "Earth",
        });

        const moon = makePlanet({
            radius: MOON_RADIUS,
            mat: {
                shininess: 5,
                map: this.system.textureLoader.load("/static/textures/planets/moon_1024.jpg")
            },
            label: "Moon"
        });

        this.system.scene.add(earth,moon);

        this.system.render = (elapsed) => {
            moon.position.set( Math.sin( elapsed ) * 5, 0, Math.cos( elapsed ) * 5 );
        }

        this.system.animate();
    }
}

const client = new Client();
client.init();

const inter = setInterval(()=>{
    if(document.getElementById("gallexy-view") !== null){
        client.initWorld();
        clearInterval(inter);
    };
},3000);


/**
 * 
 * @param {SubmitEvent} ev 
 */
function broadcastMessage(ev, target){
    const data = new FormData(ev.target);

    /**@type {HTMLFormElement} */
    const form = ev.target
    form.reset();

    const msg = data.get("msg");

    if(!target) {
        client.broadcast(msg);
        return;
    }

    client.message(target,msg);
}

/**
 * @param {MouseEvent} ev 
 */
function onInventoryClick(ev){
    /**
     * @type {HTMLElement}
     */
    const source = ev.target
    const parent = source.closest("div[data-id='item']");
    if(!parent) return;
    console.log(parent);
    
}

/**
 * @param {MouseEvent} ev 
 */
function viewMessage(ev){
    /**
     * @type {HTMLDivElement | null}
     */
    const parent = ev.target.closest("div[data-type='message']");
    if(!parent) return;

    const messageContent = parent.querySelector("p").textContent;
    if(!messageContent) return;
    const dialog = showModal("message-content")
    dialog.querySelector("p").textContent = messageContent;
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
    onInventoryClick,
    broadcastMessage,
    showModal,
    viewMessage
}