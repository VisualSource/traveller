import { Clock,TextureLoader,Scene, DirectionalLight, AxesHelper, SphereGeometry,MeshPhongMaterial,SRGBColorSpace, PerspectiveCamera, Vector2, Mesh, WebGLRenderer, Vector3, } from "three";
import { CSS2DRenderer, CSS2DObject } from 'three/addons/renderers/CSS2DRenderer.js';
/** 
 * @typedef {{ mat: import("three").MeshPhongMaterialParameters, radius?: number, position: Vector3, label: string }} PlanetOptions */


/**
 * 
 * @param {PlanetOptions} opts 
 */
export function makePlanet(opts = {}){
    const radius = opts?.radius ?? 1;
    const geomerty = new SphereGeometry(radius, 16, 16);
    const material = new MeshPhongMaterial(opts.mat);
    material.map.colorSpace = SRGBColorSpace;

    const mesh = new Mesh(geomerty,material);
    mesh.layers.enableAll();
    mesh.name = "Planet";
    mesh.userData = { name: "", id: "", active: false };

    const labelDiv = document.createElement("div");
    labelDiv.className = "label";
    labelDiv.textContent = opts.label;
    labelDiv.style.backgroundColor = "transparent";

    const clickable = document.createElement("div");
    clickable.style.borderRadius = "100%";
    clickable.style.height = `${16}px`;
    clickable.style.width = `${16}px`;
    clickable.addEventListener("click",()=>{
        console.log("Item");
    })

    const clickarea = new CSS2DObject(clickable);

    const label = new CSS2DObject(labelDiv);
    label.position.set(1.5 * radius,0,0);
    label.center.set(0,0);
    mesh.add(label,clickarea);
    label.layers.set(0);
    return mesh;
}

export class Galaxy {
    /** @type {PerspectiveCamera} */
    camera 
    clock = new Clock(); 
    scene = new Scene();
    textureLoader = new TextureLoader();
    /** @type {WebGLRenderer} */ 
    renderer
    /** @type {CSS2DRenderer} */
    labelRenderer
    /** @type {HTMLElement} */
    container
    /**
     * 
     * @param {HTMLElement} container 
     */
    constructor(container){
        this.container = container;
    }

    /** @returns {number} */
    get width(){
        return this.container?.innerWidth ?? this.container?.clientWidth ?? 0;
    }
    /** @returns {number} */
    get height(){
        return this.container?.innerHeight ?? this.container?.clientHeight ?? 0;
    }


    init(debug = false){
        this.camera = new PerspectiveCamera(45,this.width/this.height,0.1,200);
        this.camera.layers.enableAll();  
        this.scene.add(this.camera);      

        const dirLight = new DirectionalLight(0xffffff,3);
        dirLight.position.set(0,0,1);
        dirLight.layers.enableAll();
        this.scene.add(dirLight);

        if(debug){
            const axesHelper = new AxesHelper(10);
            axesHelper.layers.enableAll();
            this.scene.add(axesHelper);
        }


        this.renderer = new WebGLRenderer();
        this.renderer.setPixelRatio(window.devicePixelRatio);
        this.renderer.setSize(this.width,this.height);
        this.container.appendChild(this.renderer.domElement);

        this.labelRenderer = new CSS2DRenderer();
        this.labelRenderer.setSize(this.width,this.height);
        this.labelRenderer.domElement.style.position = "absolute";
        this.labelRenderer.domElement.style.top = "0px";

        this.container.appendChild(this.labelRenderer.domElement);
        window.addEventListener("resize",this.resize);
    } 

    resize = () => {
        this.camera.aspect = this.width / this.height;
        this.camera.updateProjectionMatrix();
        this.renderer.setSize(this.width,this.height);
        this.labelRenderer.setSize(this.width,this.height);
    }
    /** @param {number} delta */
    render = (delta) => {}

    animate = () => {
        requestAnimationFrame(this.animate);
        const elapsed = this.clock.getElapsedTime();
        this.render(elapsed);
        this.renderer.render(this.scene,this.camera);
        this.labelRenderer.render(this.scene,this.camera);
    }
}