// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {app} from '../models';

export function GetContexts():Promise<Array<string>>;

export function GetDeployments():Promise<Array<string>>;

export function GetNamespaces():Promise<Array<string>>;

export function LoadCluster(arg1:string,arg2:string):Promise<void>;

export function SetDeployment(arg1:string):Promise<Array<string>>;

export function SetNamespace(arg1:string):Promise<void>;

export function Stream():Promise<void>;

export function Test():Promise<app.PodLogMessage>;