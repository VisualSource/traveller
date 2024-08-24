function initWS(){
    const ws = new WebSocket(`ws://${document.location.host}/session/ws`);

    ws.addEventListener("close",()=>{
        console.info("Websocket has closed!");
    });
    ws.addEventListener("error",(err)=>{
        console.error(err);
    });
    ws.addEventListener("message",(msg)=>{
        console.log(msg);
    });
    ws.addEventListener("open",()=>{
        console.info("Websocket has opened!");
    });

    return ws;
}

const conn = initWS();


/**
 * Fetchs session from url
 * @returns {string}
 */
function getSessionId(){
    const path = location.pathname.split("/").filter(Boolean);
    return path.at(-1)
}

/**
 * 
 * @param {SubmitEvent} ev 
 */
function sendWS(ev){
    const data = new FormData(ev.target);
    const msg = data.get("msg")

    console.log(ev)
    conn.send(JSON.stringify({ session: getSessionId(), payload: { message: msg }, target: "global" }));
}