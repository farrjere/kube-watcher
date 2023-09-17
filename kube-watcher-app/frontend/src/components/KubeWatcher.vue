<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {GetContexts, SetDeployment, LoadCluster, GetNamespaces, GetDeployments, SetNamespace, Stream, CancelPodStream, Save, Search} from "../../wailsjs/go/app/App";
import {EventsOn} from "../../wailsjs/runtime";

import {app} from "../../wailsjs/go/models";
import PodLogMessage = app.PodLogMessage;
const logsByPod = ref(new Map<string, string>());
const count = ref(0)
const contexts = ref([""])
const sortOrder = ref("")
const namespaces = ref([""])
const deployments = ref([""])
const selectedContext = ref("")
const selectedNamespace = ref("")
const selectedDeployment = ref("")
const query = ref("")
const podNames = ref([""])
const searchOptions = ref(["Lines", "Pod Name", "Recent Update"])

onMounted(async () => {
  console.log(`The initial count is ${count.value}.`)
  contexts.value = await GetContexts();
  EventsOn("pod_log", (log_message: PodLogMessage) => {
    let podLogs = logsByPod.value.get(log_message.pod);
    if(podLogs === undefined) {
      podLogs = "";
    }

    podLogs+= log_message.message + "\n";
    logsByPod.value.set(log_message.pod, podLogs);
  })


})
function parseDate(s: string){
  var b = s.split(/\D+/);
  return new Date(+b[0], +b[1]-1, +b[2], +b[3], +b[4], +b[5]);
}

function sortPodsBySearchOption() {
  switch (sortOrder.value){
    case "Lines":
      podNames.value.sort((a, b) => {
       let aLength = 0;
       let aLogs = logsByPod.value.get(a)
       if (aLogs !== undefined) {
         aLength = aLogs.length;
       }
       let bLength = 0;
       let bLogs = logsByPod.value.get(b);
       if (bLogs !== undefined) {
         bLength = bLogs.length;
       }
        return bLength - aLength;
      });
      break;
    case "Pod Name":
      podNames.value.sort((a, b) => {
        return a.localeCompare(b);
      });
      break;
    case "Recent Update":
      podNames.value.sort((a, b) => {
        let aLogs = logsByPod.value.get(a);
        let bLogs = logsByPod.value.get(b);
        if (aLogs === undefined) {
          return 1;
        }
        if (bLogs === undefined) {
          return -1;
        }
        let aLines = aLogs.split("\n");
        let bLines = bLogs.split("\n");
        let aLastLine = aLines[aLines.length - 2];
        let bLastLine = bLines[bLines.length - 2];
        let aTimeString = aLastLine.split(" ")[0];
        let aTime = parseDate(aTimeString);
        let bTimeString   = bLastLine.split(" ")[0];
        let bTime = parseDate(bTimeString);
        console.log(aTimeString, bTimeString);
        console.log(aTime, bTime);
        return  bTime.valueOf() - aTime.valueOf();
      });
      break;

  }

}

async function stream(){
  logsByPod.value = new Map<string, string>();
  Stream();
}

async function setNamespace(){
  console.log("Called setNamespace")
  await SetNamespace(selectedNamespace.value);
  deployments.value = await GetDeployments();
}

async function cancelAllStreams() {
  console.log("Called cancelAllStreams")
  for (const name of podNames.value){
    await CancelPodStream(name);
  }
}

async function setContext() {
  console.log("Called setContext")
  await LoadCluster("", selectedContext.value);
  namespaces.value = await GetNamespaces();
}

async function setDeployment() {
  podNames.value = await SetDeployment(selectedDeployment.value);
  for (var name of podNames.value){
    logsByPod.value.set(name, "");
  }
}

async function save() {
  console.log("Called save");
  await Save();
}

async function execSearch() {
  logsByPod.value = new Map<string, string>();
  console.log("Called save");
  let searchResults = await Search(query.value, 1000);
  console.log(searchResults.length);
  for(let result of searchResults) {
    let logString = "";
    for(let m of result.matches) {
      logString += m + "\n";
    }
    logsByPod.value.set(result.pod_name, logString);
  }
}

</script>


<template class="bg-black">
  <nav class="navbar bg-dark navbar-expand-xl navbar-vertical" data-bs-theme="dark">
    <div class="container-fluid">
      <a class="navbar-brand" href="#"></a>
      <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarNavDropdown" aria-controls="navbarNavDropdown" aria-expanded="false" aria-label="Toggle navigation">
        <span class="navbar-toggler-icon"></span>
      </button>
      <div class="collapse navbar-collapse" id="navbarNavDropdown">
        <ul class="navbar-nav me-auto mb-2 mb-lg-0">
          <li class="nav-item dropdown">
            <label for="contextSelect" class="text-secondary">Context</label>
            <br/>
              <select v-model="selectedContext" @input="setContext">
                <option disabled value="">Please select a context</option>
                <option v-for="context in contexts">{{ context}}</option>
              </select>
          </li>

            <li class="nav-item dropdown">
              <label for="contextSelect" class="text-secondary">Namespace</label>
              <br/>
                <select v-model="selectedNamespace" @change="setNamespace()">
                  <option disabled value="">Please select a namespace</option>
                  <option v-for="namespace in namespaces">{{ namespace}}</option>
                </select>
            </li>
            <li class="nav-item dropdown">
              <label for="contextSelect" class="text-secondary">Deployment</label><br/>
              <select v-model="selectedDeployment" @change="setDeployment">
                <option disabled value="">Please select a deployment</option>
                <option v-for="deployment in deployments">{{ deployment}}</option>
              </select>
            </li>
          <li v-if="selectedDeployment !== ''"  class="nav-item"><p><button class="nav-item"  @click="stream()">Stream</button></p></li>
          <li v-if="podNames[0] !== ''"  class="nav-item"><p><button class="nav-item"  @click="cancelAllStreams()">Cancel All Streams</button></p></li>
          <li v-if="selectedDeployment !== ''"  class="nav-item"><p><button class="nav-item" @click="save()">Save</button></p></li>
        </ul>
            <input class="form-control me-2" style="width: 300px;" type="search" v-model="query" placeholder="Search" aria-label="Search">
            <button class="btn btn-outline-success" @click="execSearch()">Search</button>
      </div>
    </div>
  </nav>
  <div v-if="podNames[0] !== ''" class="container-fluid">
    <label class="text-secondary" for="podSort">Sort By:</label>
    <select id="podSort" @change="sortPodsBySearchOption()" v-model="sortOrder">
      <option disabled value="">Sort Order</option>
      <option v-for="searchOp in searchOptions">{{ searchOp}}</option>
    </select>
    <div class="row align-content-center">
      <div v-for="(pod, index) in podNames" class="p-1 rounded-1 text-bg-dark text-info col-lg-5 sides">
        <div class="py-5">
          <h3 class="display-5 fw-bold">{{pod}}</h3>
          <p style="white-space: pre-wrap" class="box">{{logsByPod.get(pod) }}</p>
        </div>
      </div>
    </div>
  </div>
</template>>

<style scoped>
.box {
  height: 400px;
  overflow-y: scroll;
}
.sides {
  margin-inline-start: 120px;
  margin-bottom: 5px;
  margin-top: 5px;
}
</style>