/* CS2 Demo Analyzer - progressive enhancement.
 *
 * The report is fully readable without JavaScript. This only adds navigation
 * highlighting, timeline filtering, and copying demo tick commands.
 */
(function () {
  "use strict";

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

  // Report search filters player cards, cheat cards and timeline sections by
  // their pre-computed data-search text. All content stays visible without JS.
  const searchWrap = document.getElementById("topsearch");
  const searchBox = document.getElementById("report-search");
  const searchStatus = document.getElementById("search-status");
  if (searchWrap && searchBox) {
    searchWrap.hidden = false;

    const playerCards = Array.from(document.querySelectorAll(".player-card[data-search]"));
    const cheaterCards = Array.from(document.querySelectorAll(".cheater-card[data-search]"));
    const timelineSections = Array.from(document.querySelectorAll("[data-timeline-section][data-search]"));
    const pills = Array.from(document.querySelectorAll(".player-pills .pill[data-target]"));

    const searchMatches = (el, query) => !query || (el.dataset.search || "").indexOf(query) !== -1;

    function applySearch() {
      const query = searchBox.value.trim().toLowerCase();
      const active = query.length > 0;
      let players = 0;
      let rounds = 0;

      playerCards.forEach((card) => {
        const ok = searchMatches(card, query);
        card.classList.toggle("search-hidden", !ok);
        if (ok && active) players++;
      });

      pills.forEach((pill) => {
        const card = document.getElementById(pill.dataset.target || "");
        const ok = !active || (card && !card.classList.contains("search-hidden"));
        pill.classList.toggle("search-hidden", !ok);
      });

      cheaterCards.forEach((card) => {
        card.classList.toggle("search-hidden", !searchMatches(card, query));
      });

      timelineSections.forEach((section) => {
        const ok = searchMatches(section, query);
        section.classList.toggle("search-hidden", !ok);
        if (ok && active && section.classList.contains("tl-round")) rounds++;
      });

      if (!active) {
        searchStatus.textContent = "";
        return;
      }
      const parts = [];
      if (playerCards.length) parts.push(players + (players === 1 ? " player" : " players"));
      if (timelineSections.length) parts.push(rounds + (rounds === 1 ? " round" : " rounds"));
      searchStatus.textContent = parts.join(" - ");
    }

    searchBox.addEventListener("input", applySearch);
    searchBox.addEventListener("search", applySearch);
    searchBox.addEventListener("keydown", (e) => {
      if (e.key === "Escape") {
        searchBox.value = "";
        applySearch();
        searchBox.blur();
      }
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
      return type === "bomb_plant" || type === "bomb_defuse" || type === "bomb_explode";
    }
    if (evtType === "lifecycle") {
      return type === "player_connect" || type === "player_disconnect" || type === "team_change" || type === "team_swap";
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
    if (playerFilter) {
      player = "";
      playerFilter.value = "";
      applyFilter();
    }
    if (searchBox) {
      searchBox.value = "";
      searchBox.dispatchEvent(new Event("input"));
    }
  });

  if (location.hash) {
    const target = document.querySelector(location.hash);
    if (target) {
      target.scrollIntoView({ block: "start" });
    }
  }
})();
