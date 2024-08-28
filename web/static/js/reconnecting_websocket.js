/**@typedef {CustomEvent<{ code: number, reason?: string; }>} WebsocketReconnectionEvent */
/**@typedef {CustomEvent<MessageEvent<string>>} WebsocketMessageEvent*/
/**@typedef {CustomEvent<Event>} WebsocketErrorEvent*/

/**
 * @emits open 
 * @emits error 
 * @emits message
 * @emits reconnect
 */
export class ReconnectingWebsocket extends EventTarget {
    maxReconnectionDelay = 10_000;
    minReconnectionDelay = 1000 + Math.random() * 4000
    reconnectionDelayGrowFactor = 1.3
    minUptime = 5000
    connectionTimeout = 4000
    maxRetries = Infinity
    maxEnqueuedMessges = Infinity
    startClosed = false;
    debug = false;
    #retryCount = -1;
    /**@type {number|undefined} */
    #uptimeTimeout
    /**@type {number|undefined} */
    #connectTimeout
    #shouldReconnect = true 
    #connectLock = false 
    #binaryType = "blob"
    #closeCalled = false 
    /** @type {(string | ArrayBuffer | Blob | ArrayBufferView)[]} */
    #messageQueue = []
    /**@type {string} */
    #url
    /**@type {string | string[] | undefined} */
    #protocols
    /**@type {WebSocket} */
    #ws
    /**
     * 
     * @param {string} url 
     * @param {string|string[]|undefined} protocols 
     */
    constructor(url,protocols){
        super()
        this.#url = url; 
        this.#protocols = protocols;

    
        if(this.startClosed) {
            this.#shouldReconnect = false;
        }
        this.#connect();
    }

    get binaryType(){
        return this.#ws ? this.#ws.binaryType : this.#binaryType
    }

    get retryCount(){
        return Math.max(this.#retryCount,0);
    }
    /**
     * The number of bytes of data that have been queued using calls to send() but not yet 
     * transmitted to the network. This value resets to zero once all queued data has been sent.
     * This value does not reset to zero when connections is closed; if you keep calling send().
     * @readonly
     * @returns {number}
     */
    get bufferedAmount(){
        const bytes = this.#messageQueue.reduce((acc,message)=>{
            switch (true) {
                case typeof message === "string":
                    acc += message.length
                    break;
                case message instanceof Blob: {
                    acc += message.size;
                    break;
                }
                default:
                    acc += message.byteLength;
                    break;
            }
            return acc;
        },0);
        return bytes + (this.#ws ? this.#ws.bufferedAmount : 0);
    }

    get extensions(){
        return this.#ws ? this.#ws.extensions : ""
    }

    get protocol(){
        return this.#ws ? this.#ws.protocol : ""
    }

    get readyState(){
        if(this.#ws){
            return this.#ws.readyState;
        }
        return this.startClosed ? WebSocket.CLOSED: WebSocket.CONNECTING;
    }

    get url(){
        return this.#ws ? this.#ws.url : "";
    }

    /**
     * 
     * @param {number} code 
     * @param {string | undefined} reason 
     * @returns 
     */
    close(code = 1000, reason){
        this.#closeCalled = true;
        this.#shouldReconnect = false;
        this.#clearTimeouts();
        if(!this.#ws) return;
        if(this.#ws.readyState === WebSocket.CLOSED) return;
        this.#ws.close(code,reason)
    }
    /**
     *
     * @param {*} code 
     * @param {*} reason 
     */
    reconnect(code,reason){
        this.dispatchEvent(new CustomEvent("reconnecting",{ detail: { code, reason } }));
        this.#shouldReconnect = true;
        this.#closeCalled = false;
        if(!this.#ws || this.#ws.readyState === WebSocket.CLOSED) {
            this.#connect();
            return;
        }
        this.#disconnect(code,reason);
        this.#connect();
    }

    /**
     * 
     * @param {string | ArrayBuffer | Blob | ArrayBufferView} data 
     */
    send(data){
        if(this.#ws && this.#ws.readyState === WebSocket.OPEN) {
            this.#ws.send(data);
            return;
        }
        if(this.#messageQueue.length < this.maxEnqueuedMessges) {
            this.#messageQueue.push(data);
        } 
    }

    #connect(){
        if(this.#connectLock || !this.#shouldReconnect) return;
        this.#connectLock = true;

        if(this.#retryCount >= this.maxRetries)return;

        this.#retryCount++;

        this.#removeListeners();

        this.#wait().then(()=>{
            if(this.#closeCalled) return;
            this.#ws = new WebSocket(this.#url,this.#protocols)
            this.#ws.binaryType = this.#binaryType;
            this.#connectLock = false;
            this.#addListeners();
            this.#connectTimeout = setTimeout(this.#handleTimeout,this.connectionTimeout);

        });
    }
    #handleTimeout = () => {

    }

    #clearTimeouts(){
        clearTimeout(this.#connectTimeout);
        clearTimeout(this.#uptimeTimeout);
    }
    /**
     * 
     * @param {number} code 
     * @param {string | undefined} reason 
     */
    #disconnect(code = 1000, reason){
        this.#clearTimeouts();
        if(!this.#ws) return;

        try {
            this.#ws.close(code,reason);
        } catch (error) {
            
        }
    }
    #getNextDelay(){
        let delay = 0;
        if(this.#retryCount > 0) {
            delay = this.minReconnectionDelay * Math.pow(this.reconnectionDelayGrowFactor,this.#retryCount - 1);
            if(delay > this.maxReconnectionDelay){
                delay = this.maxReconnectionDelay;
            }
        }
        return delay;
    }
    #wait(){
        return new Promise(ok=>setTimeout(ok,this.#getNextDelay()));
    }
    #acceptOpen = () => {
        this.#retryCount = 0;
    }
    #handleError = (ev) => {
        this.#disconnect(undefined,ev.message === "TIMEOUT" ? "timeout" : undefined);

        this.dispatchEvent(new CustomEvent("error",{ detail: ev }));

        this.#connect();
    }
    /**
     * 
     * @param {CloseEvent} ev 
     */
    #handleClose = (ev) => {
        this.#clearTimeouts();

        if(this.#shouldReconnect){ 
            this.#connect();
        }

        this.dispatchEvent(new CustomEvent("close",{ detail: { 
            code: ev.code, 
            reason: ev.reason,
            wasClean: ev.wasClean, 
            timeStamp: ev.timeStamp
        } }));
    }
    #handleMessage = (ev) => this.dispatchEvent(new CustomEvent("message",{ detail: ev }));
    #handleOpen = (ev) => {
        clearTimeout(this.#connectTimeout);
        this.#uptimeTimeout = setTimeout(this.#acceptOpen,this.minUptime);
        this.#ws.binaryType = this.#binaryType;

        for(const msg of this.#messageQueue){
            this.#ws.send(msg);
        }
        this.#messageQueue = [];

        this.dispatchEvent(new CustomEvent("open",{ detail: ev }));
    }
    #removeListeners(){
        if(!this.#ws) return;
        this.#ws.removeEventListener("open",this.#handleOpen);
        this.#ws.removeEventListener("close",this.#handleClose)
        this.#ws.removeEventListener("message",this.#handleMessage);
        this.#ws.removeEventListener("error",this.#handleError)
    }
    #addListeners(){
        if(!this.#ws) return;
        this.#ws.addEventListener("open",this.#handleOpen);
        this.#ws.addEventListener("close",this.#handleClose)
        this.#ws.addEventListener("message",this.#handleMessage);
        this.#ws.addEventListener("error",this.#handleError)
    }
}
