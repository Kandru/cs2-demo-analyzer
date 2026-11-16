/* CS2 Demo Analyzer - progressive enhancement.
 *
 * The report is fully readable without JavaScript. This only adds navigation
 * highlighting, the header search dropdown, timeline filtering, and copying
 * demo tick commands.
 */
(function () {
  "use strict";

  // The sticky timeline filter bar sits below the sticky section header. Measure
  // the real header height so the filter never overlaps the section title.
  const sectionHead = document.querySelector(".section-head");
  if (sectionHead) {
    const syncSectionHead = () => {
      const height = sectionHead.getBoundingClientRect().height;
      if (height > 0) {
        document.documentElement.style.setProperty("--section-head-h", height + "px");
      }
    };
    syncSectionHead();
    window.addEventListener("resize", syncSectionHead);
    if ("ResizeObserver" in window) {
      new ResizeObserver(syncSectionHead).observe(sectionHead);
    }
  }

  const navLinks = Array.from(document.querySelectorAll(".mainnav a"));
  const sections = navLinks
    .map((a) => document.querySelector(a.getAttribute("href")))
    .filter(Boolean);

  function setActive(id) {
    navLinks.forEach((a) => {
      const active = a.getAttribute("href") === "#" + id;
      a.classList.toggle("active", active);
      if (active) {
        a.setAttribute("aria-current", "page");
      } else {
        a.removeAttribute("aria-current");
      }
    });
  }

  if ("IntersectionObserver" in window) {
    const observer = new IntersectionObserver(
      (entries) => {
        const visible = entries
          .filter((e) => e.isIntersecting)
          .sort((a, b) => b.intersectionRatio - a.intersectionRatio)[0];
        if (visible) setActive(visible.target.id);
      },
      { rootMargin: "-80px 0px -60% 0px", threshold: [0.05, 0.25, 0.5] }
    );
    sections.forEach((s) => observer.observe(s));
  }

  // Header search is a combobox backed by a server-rendered list of anchors.
  // Names are read from data attributes and inserted with textContent, so
  // untrusted player/server strings never become HTML.
  const searchWrap = document.getElementById("topsearch");
  const searchBox = document.getElementById("report-search");
  const searchPop = document.getElementById("search-results");
  const searchList = document.getElementById("search-results-list");
  const searchEmpty = searchPop ? searchPop.querySelector(".search-empty") : null;

  if (searchWrap && searchBox && searchPop && searchList) {
    searchWrap.hidden = false;

    const entries = Array.from(document.querySelectorAll("#search-index a")).map((a) => ({
      href: a.getAttribute("href") || "",
      label: a.dataset.label || a.textContent || "",
      detail: a.dataset.detail || "",
      kind: a.dataset.kind || "",
      search: (a.dataset.search || "").toLowerCase(),
    }));

    let results = [];
    let active = -1;

    function close() {
      searchPop.hidden = true;
      searchList.textContent = "";
      if (searchEmpty) searchEmpty.hidden = true;
      searchBox.setAttribute("aria-expanded", "false");
      searchBox.removeAttribute("aria-activedescendant");
      active = -1;
    }

    function open() {
      searchPop.hidden = false;
      searchBox.setAttribute("aria-expanded", "true");
    }

    function select(index) {
      const items = Array.from(searchList.children);
      if (!items.length) return;
      active = (index + items.length) % items.length;
      items.forEach((item, i) => {
        item.classList.toggle("is-active", i === active);
        item.setAttribute("aria-selected", i === active ? "true" : "false");
      });
      searchBox.setAttribute("aria-activedescendant", items[active].id);
    }

    function go(entry) {
      if (!entry || !entry.href) return;
      if (entry.href.charAt(0) === "#") {
        location.hash = entry.href;
        const target = document.getElementById(entry.href.slice(1));
        if (target) target.scrollIntoView({ block: "start" });
      } else {
        location.href = entry.href;
      }
      searchBox.value = "";
      close();
      searchBox.blur();
    }

    function render(query) {
      const q = query.trim().toLowerCase();
      results = !q
        ? []
        : entries.filter((e) => e.search.indexOf(q) !== -1).slice(0, 12);
      searchList.textContent = "";
      active = -1;
      if (!q) {
        close();
        return;
      }
      if (!results.length) {
        open();
        if (searchEmpty) searchEmpty.hidden = false;
        return;
      }
      results.forEach((entry, i) => {
        const li = document.createElement("li");
        li.id = "search-opt-" + i;
        li.setAttribute("role", "option");
        li.setAttribute("aria-selected", "false");
        li.dataset.index = String(i);
        const button = document.createElement("button");
        button.type = "button";
        const label = document.createElement("b");
        label.textContent = entry.label;
        button.appendChild(label);
        if (entry.detail) {
          const detail = document.createElement("span");
          detail.textContent = entry.detail;
          button.appendChild(detail);
        }
        const kind = document.createElement("em");
        kind.textContent = entry.kind;
        button.appendChild(kind);
        li.appendChild(button);
        searchList.appendChild(li);
      });
      if (searchEmpty) searchEmpty.hidden = true;
      open();
    }

    searchBox.addEventListener("input", () => render(searchBox.value));
    searchBox.addEventListener("focus", () => {
      if (searchBox.value.trim()) render(searchBox.value);
    });
    searchBox.addEventListener("keydown", (e) => {
      if (e.key === "ArrowDown") {
        e.preventDefault();
        if (searchPop.hidden) render(searchBox.value);
        select(active + 1);
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        select(active - 1);
      } else if (e.key === "Home" && active >= 0) {
        e.preventDefault();
        select(0);
      } else if (e.key === "End" && active >= 0) {
        e.preventDefault();
        select(results.length - 1);
      } else if (e.key === "Enter") {
        if (active >= 0) {
          e.preventDefault();
          go(results[active]);
        } else if (results.length === 1) {
          e.preventDefault();
          go(results[0]);
        }
      } else if (e.key === "Escape") {
        searchBox.value = "";
        close();
        searchBox.blur();
      }
    });

    searchList.addEventListener("click", (e) => {
      const button = e.target.closest("button");
      if (!button) return;
      const li = button.closest("li");
      const index = Number(li && li.dataset.index);
      go(results[index]);
    });

    document.addEventListener("mousedown", (e) => {
      if (searchWrap.contains(e.target)) return;
      close();
    });

    document.addEventListener("keydown", (e) => {
      if (e.key !== "/" || e.ctrlKey || e.metaKey || e.altKey) return;
      const target = e.target;
      if (
        target &&
        (target.tagName === "INPUT" ||
          target.tagName === "TEXTAREA" ||
          target.tagName === "SELECT" ||
          target.isContentEditable)
      ) {
        return;
      }
      e.preventDefault();
      searchBox.focus();
      searchBox.select();
    });
  }

  const evtFilterButtons = Array.from(document.querySelectorAll("[data-evt-filter]"));
  let evtType = "all";
  const playerFilter = document.getElementById("player-filter");
  let player = "";

  function typeMatches(type) {
    if (evtType === "all") return true;
    if (evtType === "kill") return type === "kill" || type === "teamkill" || type === "suicide";
    if (evtType === "bomb") {
      return (
        type === "bomb_plant" ||
        type === "bomb_defuse" ||
        type === "bomb_explode" ||
        type === "bomb_pickup" ||
        type === "bomb_drop"
      );
    }
    if (evtType === "grenade") {
      return type === "grenade" || type === "grenade_damage" || type === "flash";
    }
    if (evtType === "lifecycle") {
      return (
        type === "player_connect" ||
        type === "player_disconnect" ||
        type === "team_change" ||
        type === "team_swap" ||
        type === "hostage_pickup" ||
        type === "hostage_drop" ||
        type === "hostage_rescue" ||
        type === "hostage_killed"
      );
    }
    if (evtType === "round") return type === "round_end" || type === "freeze_end";
    return type === evtType;
  }

  function applyFilter() {
    const hasFilter = evtType !== "all" || player !== "";
    document.querySelectorAll("[data-timeline-section]").forEach((section) => {
      let visible = 0;
      section.querySelectorAll(".tl-row").forEach((row) => {
        const players = (row.dataset.players || "").split(" ").filter(Boolean);
        const ok = typeMatches(row.dataset.type) && (!player || players.includes(player));
        row.classList.toggle("is-hidden", !ok);
        if (ok) visible++;
      });
      section.classList.toggle("is-empty", hasFilter && visible === 0);
    });
  }

  evtFilterButtons.forEach((btn) => {
    btn.addEventListener("click", () => {
      evtFilterButtons.forEach((b) => b.classList.remove("active"));
      btn.classList.add("active");
      evtType = btn.dataset.evtFilter;
      applyFilter();
    });
  });

  if (playerFilter) {
    playerFilter.addEventListener("change", () => {
      player = playerFilter.value;
      applyFilter();
    });
  }

  // Copy "demo_gototick <tick>" so the reviewer can jump in-game. Without JS
  // the tick number is still visible and selectable.
  function copyText(text) {
    if (navigator.clipboard && window.isSecureContext) {
      return navigator.clipboard.writeText(text);
    }
    const area = document.createElement("textarea");
    area.value = text;
    area.setAttribute("readonly", "");
    area.style.position = "fixed";
    area.style.left = "-9999px";
    document.body.appendChild(area);
    area.select();
    try {
      document.execCommand("copy");
    } catch (_) {}
    document.body.removeChild(area);
    return Promise.resolve();
  }

  document.querySelectorAll(".tick-copy").forEach((btn) => {
    btn.addEventListener("click", () => {
      const command = btn.dataset.copy || "";
      if (!command) return;
      copyText(command)
        .then(() => {
          const previous = btn.textContent;
          btn.textContent = "copied";
          btn.classList.add("copied");
          window.setTimeout(() => {
            btn.textContent = previous;
            btn.classList.remove("copied");
          }, 1200);
        })
        .catch(() => {});
    });
  });

  document.addEventListener("keydown", (e) => {
    if (e.key !== "Escape") return;
    if (playerFilter && player) {
      player = "";
      playerFilter.value = "";
      applyFilter();
    }
  });

  if (location.hash) {
    try {
      const target = document.querySelector(location.hash);
      if (target) target.scrollIntoView({ block: "start" });
    } catch (_) {
      // Ignore malformed user-entered fragments.
    }
  }
})();
