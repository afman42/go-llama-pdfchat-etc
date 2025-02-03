import "./style.css";
import "select2/select2.css";
import "toastr/build/toastr.css";
import "jquery/jquery";
import "select2/select2";
import toastr from "toastr/toastr";

//Lib
toastr.options.showEasing = "swing";
toastr.options.hideEasing = "linear";
toastr.options.closeEasing = "linear";
toastr.options.closeButton = true;
toastr.options.closeMethod = "fadeOut";
toastr.options.closeDuration = 1000;
toastr.options.preventDuplicates = true;
toastr.options.timeOut = 1000; // How long the toast will display without user interaction
toastr.options.extendedTimeOut = 1000; // How long the toast will display after a user hovers over it
toastr.options.progressBar = true;
$("#app").html(
  `<main class="flex mx-auto flex-col m-2 w-96">
    <div class="flex border-2 gap-2 px-2 py-2 pb-2 flex-col border-black border-solid w-full h-60 mb-2 overflow-y-auto" id="txtContainer">
    </div>
    <div class="flex flex-col items-center justify-center w-full">
      <textarea disabled id="txt" class="focus:border-blue-500 mb-2 flex w-full border-solid border-2 border-black rounded-lg p-2" name="txt" placeholder="Hello World">
      </textarea>
      <div class="grid grid-cols-2 gap-2 flex-row w-full">
        <form class="flex w-full cursor-pointer" id="formFile">
          <input type="file" id="file" name="file" class="w-full p-2  bg-blue-100 rounded-lg hover:border-solid hover:border-2 hover:border-blue-100 hover:bg-white hover:text-black cursor-not-allowed">
        </form>
        <button type="button" disabled id="send" class="flex w-full p-2 bg-red-500 cursor-not-allowed rounded-lg hover:bg-white hover:border-solid hover:border-red-500 hover:border-2 hover:text-black text-white justify-center">send</button>
      </div>
    </div>
    <div class="flex flex-col my-2">
      <label for="remote-url" class="block">Fetch Models
          <button type="button" id="btnHtUrl" class="flex w-full py-2 bg-amber-500 cursor-pointer rounded-lg hover:bg-white hover:border-solid hover:border-amber-500 hover:border-2 hover:text-black text-white justify-center items-center">Fetch</button>
      </label>
      <div class="grid grid-cols-2 gap-2 flex-row w-full">
        <label for="embedding" class="flex flex-col">Model Embbeding <a href="https://ollama.com/search?c=embedding" target="_blank" class="text-blue-500">Link models</a>
          <input type="hidden" class="basic-single-embed"/>
        </label>
        <label for="chat" class="flex flex-col">Model Chat <a href="https://ollama.com/search" target="_blank" class="text-blue-500">Link Chat</a>
          <input type="hidden" class="basic-single-chat"/>
        </label>
      </div>
    </div>
    <span>Log:</span>
    <div class="flex border-2 border-black border-solid w-full h-48 overflow-x-scroll " id="logStdout">
    </div>
  </main>`,
);

const formData = new FormData();
const API_URL = import.meta.env.VITE_API_URL;
let data = {};
$("#txt").addClass("cursor-not-allowed");
$("#file").attr("disabled", true);
$("#file").on("change", function (e) {
  let fileList = e.target.files;
  if (!fileList.length) return;
  formData.append("file", fileList[0], fileList[0].name);
  $("#send").attr("disabled", false);
  $("#send").addClass("cursor-pointer");
  $("#txt").removeClass("cursor-not-allowed");
  $("#txt").attr("disabled", false);
});

let dataModelsArray = [];
$(".basic-single-embed").select2({
  placeholder: "Search for a Models Embed",
  data: [],
});

$(".basic-single-chat").select2({
  placeholder: "Search for a Models Chat",
  data: [],
});

$(".basic-single-embed").attr("disabled", true);
$(".basic-single-chat").attr("disabled", true);

function ajaxFetchSelect2() {
  $.ajax({
    url: API_URL + "/?listModel=all",
    method: "post",
    dataType: "json",
    beforeSend: function (jqXHR, settings) {
      $(".basic-single-embed").attr("disabled", true);
      $(".basic-single-chat").attr("disabled", true);
      $("#btnHtUrl").attr("disabled", true);
      toastr.info("info fetch models");
    },
    complete: function (jqXHR, txtStatus) {
      $("#btnHtUrl").attr("disabled", false);
    },
    success: function (data, textStatus, jqXHR) {
      $("#btnHtUrl").attr("disabled", false);
      const jsonParseMessage = JSON.parse(data.message.trim());
      console.log(jsonParseMessage);
      if (Array.isArray(jsonParseMessage)) {
        toastr.success("successfully response fetch");
        $(".basic-single-embed").attr("disabled", false);
        $(".basic-single-chat").attr("disabled", false);
        jsonParseMessage.forEach(function (v) {
          dataModelsArray.push({ id: v.model, text: v.model });
        });
      }
    },
    error: function (jqXHR, textStatus, errorThrown) {
      console.log({ jqXHR, toastr, errorThrown });
      toastr.error("Something went wrong fetch ajax");
    },
  });
}
$("#btnHtUrl").on("click", function (e) {
  e.preventDefault();
  ajaxFetchSelect2();

  $(".basic-single-embed").select2({
    placeholder: "Search for a Models Embed",
    data: dataModelsArray,
  });
  $(".basic-single-chat").select2({
    placeholder: "Search for a Models Chat",
    data: dataModelsArray,
  });
  formData.append("htUrl", "");
  $("#file").addClass("cursor-pointer");
  $("#file").removeClass("cursor-not-allowed");
  $("#file").attr("disabled", false);
});

$(".basic-single-embed").on("change", function (e) {
  const value = e.target.value;
  formData.append("modelEmbed", value);
});
$(".basic-single-chat").on("change", function (e) {
  const value = e.target.value;
  formData.append("modelChat", value);
});

$("#txt").on("input", function (e) {
  const value = e.target.value;
  data.txt = value;
  formData.append("txt", value);
});

$("#send").on("click", function (e) {
  e.preventDefault();
  if (data.txt == "") return toastr.warning("please fill the input");
  $("#txtContainer").append(
    '<h6 class="flex bg-gray-100 w-max px-2 py-1 rounded-lg">User: ' +
      data.txt +
      "</h6>",
  );
  $.ajax({
    url: API_URL,
    method: "post",
    dataType: "json",
    data: formData,
    cache: false,
    processData: false,
    contentType: false,
    beforeSend: function (jqXHR, settings) {
      $(".basic-single-embed").attr("readonly", true);
      $(".basic-single-chat").attr("readonly", true);
      $("#btnHtUrl").attr("disabled", true);
      $("#send").attr("disabled", true);
      $("#txt").attr("readonly", true);
      $("#file").attr("readonly", true);

      toastr.info("Info fetching", {
        timeOut: 6000,
        extendedTimeOut: 1000,
      });
    },
    complete: function (jqXHR, txtStatus) {
      $("#btnHtUrl").attr("disabled", false);
    },
    success: function (data, textStatus, jqXHR) {
      $("#btnHtUrl").attr("disabled", false);
      $(".basic-single-embed").attr("readonly", false);
      $(".basic-single-chat").attr("readonly", false);
      $("#btnHtUrl").attr("disabled", false);
      $("#send").attr("disabled", false);
      $("#txt").attr("readonly", false);
      $("#file").attr("readonly", false);
      toastr.success("successfully response fetch");
      $("#txtContainer").append(
        '<h6 class="flex bg-amber-100 w-max px-2 py-1 rounded-lg">AI: ' +
          data.message +
          "</h6>",
      );
    },
    error: function (jqXhr, textStatus, errorThrown) {
      console.log({ jqXhr, textStatus, errorThrown });
      if (textStatus == "error") {
        const jsonParseMessage = JSON.parse(jqXhr.responseText.trim());
        if (jsonParseMessage.statusCode == 400)
          toastr.error(jsonParseMessage.message);
      }
    },
  });
});
