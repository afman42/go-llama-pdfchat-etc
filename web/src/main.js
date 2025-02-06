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
  `<main class="flex mx-auto flex-col m-2 w-1/2 min-sm:w-5/6 sm:w-5/6 md:5/6">
    <div class="flex border-2 gap-2 p-2 pb-2 flex-col border-black border-solid w-full overflow-x-hidden overflow-y-auto h-60 mb-2" id="txtContainer">
    </div>
    <div class="flex flex-col items-center justify-center w-full">
      <form class="flex flex-col w-full" id="formUpload">
        <textarea value="Gollama" disabled id="txt" class="focus:border-blue-500 mb-2 flex w-full border-solid border-2 border-black rounded-lg p-2" name="txt" placeholder="Gollama"></textarea>
        <div class="grid grid-cols-2 gap-2 flex-row w-full">
          <input accept=".pdf,.txt" type="file" id="file" name="file" class="w-full p-2  bg-blue-100 rounded-lg hover:border-solid hover:border-2 hover:border-blue-100 hover:bg-white hover:text-black cursor-not-allowed">
          <button class='flex bg-green-500 text-white w-full p-2 rounded-lg hover:border-solid hover:border-2 hover:border-green-100 hover:bg-white hover:text-black' id='rmvfl' type='button'>Remove file</button>
          <button type="submit" disabled id="send" class="flex w-full p-2 bg-red-500 cursor-not-allowed rounded-lg hover:bg-white hover:border-solid hover:border-red-500 hover:border-2 hover:text-black text-white justify-center">send</button>
        </div>
      </form>
    </div>
    <div class="flex flex-col my-2">
      <label for="remote-url" class="block min-sm:text-sm">Fetch Models
          <button type="button" id="btnHtUrl" class="flex w-full py-2 bg-amber-500 cursor-pointer rounded-lg hover:bg-white hover:border-solid hover:border-amber-500 hover:border-2 hover:text-black text-white justify-center items-center min-sm:text-sm">Fetch</button>
      </label>
      <div class="grid grid-cols-2 gap-2 flex-row w-full">
        <label for="embedding" class="flex flex-col">
          <h6 class="flex flex-row my-2 min-sm:block min-sm:text-xs"><a href="https://ollama.com/search?c=embedding" target="_blank" class="text-blue-500 ml-2">Link models Embed</a></h6>
          <input type="hidden" class="basic-single-embed flex"/>
        </label>
        <label for="chat" class="flex flex-col">
          <span class="flex flex-row my-2 min-sm:block min-sm:text-xs"><a href="https://ollama.com/search" target="_blank" class="ml-2 flex text-blue-500">Link models Chat</a></span>
          <input type="hidden" class="basic-single-chat flex"/>
        </label>
      </div>
    </div>
    <span class="min-sm:text-sm">Log:</span>
    <div class="flex border-2 border-black border-solid w-full h-48 overflow-y-scroll flex-col" id="logStdout">
    </div>
  </main>`,
);

const formData = new FormData();
const API_URL = import.meta.env.VITE_API_URL;
let data = {};
data.txt = $("#txt").val();
data.modelChat = "";
data.modelEmbed = "";
data.ModelsArray = [];
data.stdout = "";
$("#txt").addClass("cursor-not-allowed");
$("#file").attr("disabled", true);
$("#rmvfl").hide();
$("#file").on("change", function (e) {
  let fileList = e.target.files;
  if (data.modelChat == "" || data.modelChat == "") {
    toastr.warning("please fill model embed or chat");
    $(this).val("");
    return;
  }
  if (!fileList.length) return;
  if (!["application/pdf", "text/plain"].includes(fileList[0].type)) {
    toastr.warning("upload file must extenstion .pdf and .txt");
    $(this).val("");
    return;
  }
  if (fileList[0].size / 1024 / 1024 > 1) {
    toastr.warning("upload file must 1024kb");
    $(this).val("");
    return;
  }
  formData.append("file", fileList[0], fileList[0].name);
  $("#send").attr("disabled", false);
  $("#send").addClass("cursor-pointer");
  $("#txt").removeClass("cursor-not-allowed");
  $("#txt").attr("disabled", false);
  $("#file").parent().removeClass("grid-cols-2");
  $("#file").parent().addClass("grid-cols-3");
  $("#file").addClass("cursor-not-allowed");
  $("#rmvfl").show();
  $("#rmvfl").addClass("cursor-pointer");
  $("#file").attr("disabled", true);
});

$("body").on("click", "#rmvfl", function (e) {
  e.preventDefault();
  formData.delete("file");
  $("#file").val("");
  $("#file").attr("disabled", false);
  $("#file").removeClass("cursor-not-allowed");
  $("#file").addClass("cursor-pointer");
  toastr.success("successfully remove file");
  $("#file").parent().removeClass("grid-cols-3");
  $("#file").parent().addClass("grid-cols-2");
  $("#txt").attr("disabled", true);
  $("#send").attr("disabled", true);
  $("#send").addClass("cursor-not-allowed");
  $("#txt").addClass("cursor-not-allowed");
  $("#rmvfl").hide();
});

$(".basic-single-embed").select2({
  placeholder: "Search for a Models Embed",
  data: data.ModelsArray,
});

$(".basic-single-chat").select2({
  placeholder: "Search for a Models Chat",
  data: data.ModelsArray,
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
      $("#btnHtUrl").addClass("cursor-not-allowed");
      $("#btnHtUrl").removeClass("cursor-pointer");
    },
    complete: function (jqXHR, txtStatus) {
      $("#btnHtUrl").attr("disabled", false);
      $("#btnHtUrl").removeClass("cursor-not-allowed");
      $("#btnHtUrl").addClass("cursor-pointer");
    },
    success: function (dt, textStatus, jqXHR) {
      $("#btnHtUrl").attr("disabled", false);
      $("#btnHtUrl").removeClass("cursor-not-allowed");
      $("#btnHtUrl").addClass("cursor-pointer");
      const jsonParseMessage = JSON.parse(dt.message.trim());
      if (Array.isArray(jsonParseMessage)) {
        toastr.success("successfully response fetch");
        $(".basic-single-embed").attr("disabled", false);
        $(".basic-single-chat").attr("disabled", false);
        jsonParseMessage.forEach(function (v) {
          data.ModelsArray.push({ id: v.model, text: v.model });
        });
      }
    },
    error: function (jqXHR, textStatus, errorThrown) {
      console.log({ jqXHR, toastr, errorThrown });
      toastr.error("Something went wrong fetch ajax");
    },
  });
}
$("body").on("click", "#btnHtUrl", function (e) {
  e.preventDefault();
  ajaxFetchSelect2();
  $(".basic-single-embed").select2({
    placeholder: "Search for a Models Embed",
    data: data.ModelsArray,
  });
  $(".basic-single-chat").select2({
    placeholder: "Search for a Models Chat",
    data: data.ModelsArray,
  });
  $("#file").addClass("cursor-pointer");
  $("#file").removeClass("cursor-not-allowed");
  $("#file").attr("disabled", false);
});

$(".basic-single-embed").on("change", function (e) {
  const value = e.target.value;
  if (formData.has("modelEmbed")) {
    formData.delete("modelEmbed");
  }
  data.modelEmbed = value;
  formData.append("modelEmbed", value);
});
$(".basic-single-chat").on("change", function (e) {
  const value = e.target.value;
  if (formData.has("modelChat")) {
    formData.delete("modelChat");
  }
  data.modelChat = value;
  formData.append("modelChat", value);
});

$("#txt").on("input", function (e) {
  const value = e.target.value;
  data.txt = value;
});

$("#formUpload").on("submit", function (e) {
  e.preventDefault();
  if (data.txt == "") {
    toastr.warning("please fill the input");
    return;
  }

  if (data.modelEmbed == "" || data.modelChat == "") {
    toastr.warning("please fill model embed or chat");
    return;
  }

  formData.append("txt", data.txt);
  $("#txtContainer").append(
    '<h6 class="flex bg-gray-100 w-full px-2 py-1 rounded-lg break-all">' +
      data.txt +
      " - 👤" +
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
      $("body *").attr("disabled", "disabled").off("click");
      $(".basic-single-embed").attr("disabled", true);
      $(".basic-single-chat").attr("disabled", true);
      $("#btnHtUrl").attr("disabled", true);
      $("#send").attr("disabled", true);
      $("#txt").attr("disabled", true);
      toastr.info("Info fetching");
      $("body").css("opacity", 0.5);
      $("body").css("cursor", "not-allowed");
      $("body").css("cursor", "not-allowed");
      $("#file").attr("disabled", true);
      $("#rmvfl").attr("disabled", true);
      $("body").css(
        "background-image",
        `url("data:image/svg+xml,%3Csvg width='24' height='24' stroke='%23000' viewBox='0 0 24 24' xmlns='http://www.w3.org/2000/svg'%3E%3Cstyle%3E.spinner_V8m1%7Btransform-origin:center;animation:spinner_zKoa 2s linear infinite%7D.spinner_V8m1 circle%7Bstroke-linecap:round;animation:spinner_YpZS 1.5s ease-in-out infinite%7D@keyframes spinner_zKoa%7B100%25%7Btransform:rotate(360deg)%7D%7D@keyframes spinner_YpZS%7B0%25%7Bstroke-dasharray:0 150;stroke-dashoffset:0%7D47.5%25%7Bstroke-dasharray:42 150;stroke-dashoffset:-16%7D95%25,100%25%7Bstroke-dasharray:42 150;stroke-dashoffset:-59%7D%7D%3C/style%3E%3Cg class='spinner_V8m1'%3E%3Ccircle cx='12' cy='12' r='9.5' fill='none' stroke-width='3'%3E%3C/circle%3E%3C/g%3E%3C/svg%3E%0A"`,
      );
      $("body").css("background-size", "20%");
      $("body").css("background-repeat", "no-repeat");
      $("body").css("background-position", "50% 50%");
      $("body").css("background-attachment", "fixed");
    },
    complete: function (jqXHR, txtStatus) {
      $("body").removeAttr("style");
      $("main").removeAttr("style");
      $("body *").removeAttr("disabled");
      $("#rmvfl").attr("disabled", false);
      $("#btnHtUrl").attr("disabled", false);
      if (formData.has("txt")) {
        formData.delete("txt");
      }

      $("#txtContainer").scrollTop($("#txtContainer")[0].scrollHeight);
    },
    success: function (data, textStatus, jqXHR) {
      $("body *").removeAttr("disabled");
      $("body").removeAttr("style");
      $("main").removeAttr("style");
      $("#btnHtUrl").attr("disabled", false);
      $(".basic-single-embed").attr("disabled", false);
      $(".basic-single-chat").attr("disabled", false);
      $("#btnHtUrl").attr("disabled", false);
      $("#send").attr("disabled", false);
      $("#txt").attr("disabled", false);
      $("#rmvfl").attr("disabled", false);
      toastr.success("successfully response fetch");
      $("#txtContainer").append(
        '<h6 class="flex bg-amber-100 w-full px-2 py-1 rounded-lg break-all">' +
          data.message +
          "- 🤖" +
          "</h6>",
      );

      $("#txtContainer").scrollTop($("#txtContainer")[0].scrollHeight);
    },
    error: function (jqXhr, textStatus, errorThrown) {
      console.log({ jqXhr, textStatus, errorThrown });
      if (textStatus == "error") {
        $("body *").removeAttr("disabled");
        $("#txt").attr("disabled", false);
        $("#rmvfl").attr("disabled", false);
        $("main").removeAttr("style");
        $("body").removeAttr("style");
        $("#btnHtUrl").attr("disabled", false);
        const jsonParseMessage = JSON.parse(jqXhr.responseText.trim());
        if (jsonParseMessage.statusCode == 400)
          toastr.error(jsonParseMessage.message);
        if (jsonParseMessage.statusCode == 500)
          toastr.error(jsonParseMessage.message);
      }
    },
  });
});
console.log(import.meta.env.MODE);
let wsport =
  import.meta.env.MODE != "staging" && import.meta.env.MODE != "development"
    ? "wss://"
    : "ws://";
var ws = new WebSocket(wsport + API_URL.split("//")[1] + "/ws");
ws.onopen = function (event) {
  console.log("Connection is open ...");
};

ws.onerror = function (err) {
  console.log("err: ", err, err.toString());
};

// Event handler for receiving text from the server
ws.onmessage = function (event) {
  if (data.stdout != event.data) {
    data.stdout = data.stdout + event.data + "\n";
  }
  let txt = data.stdout.split(/\n/).map(function (line) {
    return "<h6 class='flex p-1'>" + line + "</line>";
  });
  console.log("Received: " + event.data);
  $("#logStdout").html(txt);
  $("#logStdout").scrollTop($("#logStdout")[0].scrollHeight);
};

ws.onclose = function () {
  console.log("Connection is closed...");
};
