/* CS2 Demo Analyzer — progressive enhancement. The report is fully readable
   without JavaScript; this only adds navigation, search, filtering, folding,
   and copying demo tick commands. */
(function () {
  "use strict";

  const navLinks = Array.from(document.querySelectorAll(".mainnav a"));
  const sections = navLinks
    .map((a) => document.querySelector(a.getAttribute("href")))
    .filter(Boolean);

  function setActive(id) {
    navLinks.forEach((a) => a.classList.toggle("active", a.getAttribute("href") === "#" + id));
  }

  function openHashTarget() {
    if (!location.hash) return;
    let target = null;
    try {
      target = document.querySelector(location.hash);
    } catch (_) {
      return;
    }
    if (target && target.tagName === "DETAILS") target.open = true;
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

  // Keep the active tab in sync immediately, even before the observer fires.
  navLinks.forEach((a) => {
    a.addEventListener("click", () => {
      const href = a.getAttribute("href") || "";
      if (href.charAt(0) === "#") setActive(href.slice(1));
    });
  });
  window.addEventListener("hashchange", () => {
    const id = location.hash.replace(/^#/, "");
    if (id) setActive(id);
    openHashTarget();
  });

  document.querySelectorAll("details.round").forEach((d) => {
    d.addEventListener("toggle", () => {
      if (d.open) {
        try {
          history.replaceState(null, "", "#" + d.id);
        } catch (_) {}
      }
    });
  });

  const chips = Array.from(document.querySelectorAll("#round-chips .chip"));
  chips.forEach((chip) => {
    chip.addEventListener("click", () => {
      chips.forEach((c) => c.classList.remove("active"));
      chip.classList.add("active");
    });
  });

  const evtFilterButtons = Array.from(document.querySelectorAll("[data-evt-filter]"));
  let evtType = "all";
  const playerFilter = document.getElementById("player-filter");
  let player = "";

  function typeMatches(type) {
    if (evtType === "all") return true;
    if (evtType === "kill") return type === "kill" || type === "teamkill" || type === "suicide";
    if (evtType === "bomb") return type === "bomb_plant" || type === "bomb_defuse" || type === "bomb_explode";
    return type === evtType;
  }

  function applyFilter() {
    document.querySelectorAll("details.round").forEach((round) => {
      let visible = 0;
      round.querySelectorAll(".evt").forEach((evt) => {
        const players = (evt.dataset.players || "").split(" ").filter(Boolean);
        const ok = typeMatches(evt.dataset.type) && (!player || players.includes(player));
        evt.classList.toggle("is-hidden", !ok);
        if (ok) visible++;
      });
      const hasFilter = evtType !== "all" || player !== "";
      round.classList.toggle("filtered-empty", hasFilter && visible === 0);
      if (hasFilter && visible > 0) round.open = true;
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

  const expand = document.getElementById("expand-rounds");
  const collapse = document.getElementById("collapse-rounds");
  if (expand) expand.addEventListener("click", () => document.querySelectorAll("details.round").forEach((d) => (d.open = true)));
  if (collapse) collapse.addEventListener("click", () => document.querySelectorAll("details.round").forEach((d) => (d.open = false)));

  // Report search: filters player cards, cheat watch cards and rounds by their
  // pre-computed data-search text. All content stays visible without JS.
  const searchWrap = document.getElementById("topsearch");
  const searchBox = document.getElementById("report-search");
  const searchStatus = document.getElementById("search-status");
  if (searchWrap && searchBox) {
    searchWrap.hidden = false;

    const playerCards = Array.from(document.querySelectorAll(".player-card[data-search]"));
    const cheaterCards = Array.from(document.querySelectorAll(".cheater-card[data-search]"));
    const roundPanels = Array.from(document.querySelectorAll("details.round[data-search]"));
    const pills = Array.from(document.querySelectorAll(".player-pills .pill[data-target]"));
    const roundChips = Array.from(document.querySelectorAll("#round-chips .chip[data-target]"));

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

      roundPanels.forEach((panel) => {
        const ok = searchMatches(panel, query);
        panel.classList.toggle("search-hidden", !ok);
        if (ok && active) {
          rounds++;
          panel.open = true;
        }
      });

      roundChips.forEach((chip) => {
        const panel = document.getElementById(chip.dataset.target || "");
        const ok = !active || (panel && !panel.classList.contains("search-hidden"));
        chip.classList.toggle("search-hidden", !ok);
      });

      if (!active) {
        searchStatus.textContent = "";
        return;
      }
      const parts = [];
      if (playerCards.length) parts.push(players + (players === 1 ? " player" : " players"));
      if (roundPanels.length) parts.push(rounds + (rounds === 1 ? " round" : " rounds"));
      searchStatus.textContent = parts.join(" · ");
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

  openHashTarget();
})();
