// MITRERedTeam 开发文档交互脚本

(function () {
  "use strict";

  // 依据当前滚动位置高亮侧栏导航中对应的章节链接。
  const navLinks = Array.from(document.querySelectorAll(".sidebar nav a"));
  const sections = navLinks
    .map(function (link) {
      const id = link.getAttribute("href");
      if (!id || id.charAt(0) !== "#") {
        return null;
      }
      return document.querySelector(id);
    })
    .filter(Boolean);

  function onScroll() {
    const offset = 120;
    let currentId = null;
    for (let i = 0; i < sections.length; i++) {
      if (sections[i].getBoundingClientRect().top <= offset) {
        currentId = sections[i].id;
      }
    }
    navLinks.forEach(function (link) {
      const target = link.getAttribute("href");
      link.classList.toggle("active", target === "#" + currentId);
    });
  }

  window.addEventListener("scroll", onScroll, { passive: true });
  onScroll();
})();
