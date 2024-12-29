"use strict";

document.addEventListener("DOMContentLoaded", function (event) {
  const passwdBtn = document.getElementById("password");
  let visible = false;
  passwdBtn.onclick = function () {
    visible = !visible;
    if (visible) {
      passwdBtn.setAttribute("type", "text");
    } else {
      passwdBtn.setAttribute("type", "password");
    }
  };
});
