import "./style.css";
import "select2/select2.css";
import "jquery/jquery";
import "select2/select2";

$("#app").html(
  `<main class="flex mx-auto flex-col m-2 w-96">
    <div class="flex border-2 gap-2 px-2 pb-2 flex-col border-black border-solid w-full h-60 mb-2 overflow-y-auto">
      <h6 class="flex bg-gray-100 w-max px-2 py-1 rounded-lg">Text</h6>
    </div>
    <div class="flex flex-col items-center justify-center w-full">
      <textarea disabled id="txt" class="focus:border-blue-500 mb-2 flex w-full border-solid border-2 border-black rounded-lg p-2" name="txt" placeholder="Hello World">
Hello World
      </textarea>
      <div class="grid grid-cols-2 gap-2 flex-row w-full">
        <form class="flex w-full cursor-pointer" id="formFile">
          <input type="file" id="file" name="file" class="w-full p-2 r cursor-pointer bg-blue-100 rounded-lg hover:border-solid hover:border-2 hover:border-blue-100 hover:bg-white hover:text-black">
        </form>
        <button type="button" disabled id="send" class="flex w-full p-2 bg-red-500 cursor-not-allowed rounded-lg hover:bg-white hover:border-solid hover:border-red-500 hover:border-2 hover:text-black text-white justify-center">send</button>
      </div>
    </div>
    <div class="flex flex-col my-2">
      <label for="remote-url" class="flex flex-col">Remote URL or localhost
        <form method="get" class="flex flex-row h-12 gap-2" id="formHtUrl">
          <input id="urlRemote" class="w-64 px-2 focus:border-blue-500 mb-2 flex border-solid border-2 border-black rounded-lg" id="hUrl" value="http://localhost:11434" placeholder="http://localhost:11434" />
          <button type="submit" id="btnHtUrl" class="flex w-24 bg-amber-500 cursor-pointer rounded-lg hover:bg-white hover:border-solid hover:border-amber-500 hover:border-2 hover:text-black text-white justify-center items-center">Fetch</button>
        </form>
      </label>
      <div class="grid grid-cols-2 gap-2 flex-row w-full">
        <label for="embedding" class="flex flex-col">Model Embbeding <a href="https://ollama.com/search?c=embedding" target="_blank" class="text-blue-500">Link models</a>
          <input type="hidden" class="basic-single"/>
        </label>
        <label for="chat" class="flex flex-col">Model Chat <a href="https://ollama.com/search" target="_blank" class="text-blue-500">Link Chat</a>
          <input type="hidden" class="basic-single"/>
        </label>
      </div>
    </div>
    <span>Log:</span>
    <div class="flex border-2 border-black border-solid w-full h-48 overflow-x-scroll " id="logStdout">
    </div>
  </main>`,
);

const formData = new FormData();
$("#txt").addClass("cursor-not-allowed");
$("#file").on("change", function (e) {
  let fileList = e.target.files;
  if (!fileList.length) return;
  formData.append("name", fileList[0], fileList[0].name);
  $("#send").attr("disabled", false);
  $("#send").addClass("cursor-pointer");
  $("#txt").removeClass("cursor-not-allowed");
  $("#txt").attr("disabled", false);
});

let dataModelsArray = [];
$(".basic-single").select2({
  placeholder: "Search for a Models",
  data: [],
});

$(".basic-single").attr("disabled", true);

function ajaxFetchSelect2(urlHttp) {
  $.ajax({
    url: urlHttp + "/api/tags",
    method: "get",
    dataType: "json",
    beforeSend: function (jqXHR, settings) {
      $(".basic-single").attr("disabled", true);
      $("#btnHtUrl").attr("disabled", true);
    },
    complete: function (jqXHR, txtStatus) {
      $("#btnHtUrl").attr("disabled", false);
    },
    success: function (data, textStatus, jqXHR) {
      $("#btnHtUrl").attr("disabled", false);
      if (Array.isArray(data.models)) {
        $(".basic-single").attr("disabled", false);
        data.models.forEach(function (v) {
          dataModelsArray.push({ id: v.name, text: v.name });
        });
      }
    },
  });
}

$("#formHtUrl").on("submit", function (e) {
  e.preventDefault();
  const urlHttp = e.target[0].value;
  ajaxFetchSelect2(urlHttp);
  $(".basic-single").select2({
    placeholder: "Search for a Models",
    data: dataModelsArray,
  });
});
