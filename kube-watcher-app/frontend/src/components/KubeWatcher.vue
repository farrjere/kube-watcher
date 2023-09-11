<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {GetContexts, SetDeployment, LoadCluster, GetNamespaces, GetDeployments, SetNamespace, Stream} from "../../wailsjs/go/app/App";
import {EventsOn} from "../../wailsjs/runtime";

import {app} from "../../wailsjs/go/models";
import PodLogMessage = app.PodLogMessage;
const logsByPod = ref(new Map<string, string>());
const count = ref(0)
const contexts = ref([""])
const namespaces = ref([""])
const deployments = ref([""])
const selectedContext = ref("")
const selectedNamespace = ref("")
const selectedDeployment = ref("")
const podNames = ref([""])


onMounted(async () => {
  console.log(`The initial count is ${count.value}.`)
  contexts.value = await GetContexts();
  EventsOn("pod_log", (log_message: PodLogMessage) => {
    let podLogs = logsByPod.value.get(log_message.pod);
    if(podLogs === undefined) {
      podLogs = "";
    }
    console.log("Adding a message")
    podLogs+= log_message.message + "\n";
    logsByPod.value.set(log_message.pod, podLogs);
  })


})

async function stream(){
  Stream();
}

async function setNamespace(){
  console.log("Called setNamespace")
  await SetNamespace(selectedNamespace.value);
  deployments.value = await GetDeployments();
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

</script>


<template class="bg-dark">
  <nav class="navbar bg-dark navbar-expand-xl navbar-vertical" data-bs-theme="dark">
    <div class="container-fluid">
      <a class="navbar-brand" href="#">Navbar</a>
      <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarNavDropdown" aria-controls="navbarNavDropdown" aria-expanded="false" aria-label="Toggle navigation">
        <span class="navbar-toggler-icon"></span>
      </button>
      <div class="collapse navbar-collapse" id="navbarNavDropdown">
        <ul class="navbar-nav me-auto mb-2 mb-lg-0">
          <li class="nav-item dropdown">
            <label for="contextSelect" class="">Context</label>
            <br/>
              <select v-model="selectedContext" @input="setContext">
                <option disabled value="">Please select a context</option>
                <option v-for="context in contexts">{{ context}}</option>
              </select>
          </li>

            <li class="nav-item dropdown">
              <label for="contextSelect" class="">Namespace</label>
              <br/>
                <select v-model="selectedNamespace" @change="setNamespace">
                  <option disabled value="">Please select a namespace</option>
                  <option v-for="namespace in namespaces">{{ namespace}}</option>
                </select>
            </li>
            <li class="nav-item dropdown">
              <label for="contextSelect" class="">Deployment</label><br/>
              <select v-model="selectedDeployment" @change="setDeployment">
                <option disabled value="">Please select a deployment</option>
                <option v-for="deployment in deployments">{{ deployment}}</option>
              </select>
            </li>
          <li v-if="selectedDeployment !== ''"  class="nav-item"><p><button class="nav-item"  @click="stream()">Stream</button></p></li>
          <li v-if="selectedDeployment !== ''"  class="nav-item"><p><button class="nav-item" @click="">Save</button></p></li>
        </ul>
            <form v-if="selectedDeployment !== ''" class="d-flex" role="search">
              <input class="form-control me-2" type="search" placeholder="Search" aria-label="Search">
              <button class="btn btn-outline-success" type="submit">Search</button>
            </form>
      </div>
    </div>
  </nav>
  <div v-if="podNames.length > 0" class="container-fluid">
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