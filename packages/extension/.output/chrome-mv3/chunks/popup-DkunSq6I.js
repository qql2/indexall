(function polyfill() {
  const relList = document.createElement("link").relList;
  if (relList && relList.supports && relList.supports("modulepreload")) {
    return;
  }
  for (const link of document.querySelectorAll('link[rel="modulepreload"]')) {
    processPreload(link);
  }
  new MutationObserver((mutations) => {
    for (const mutation of mutations) {
      if (mutation.type !== "childList") {
        continue;
      }
      for (const node of mutation.addedNodes) {
        if (node.tagName === "LINK" && node.rel === "modulepreload")
          processPreload(node);
      }
    }
  }).observe(document, { childList: true, subtree: true });
  function getFetchOpts(link) {
    const fetchOpts = {};
    if (link.integrity) fetchOpts.integrity = link.integrity;
    if (link.referrerPolicy) fetchOpts.referrerPolicy = link.referrerPolicy;
    if (link.crossOrigin === "use-credentials")
      fetchOpts.credentials = "include";
    else if (link.crossOrigin === "anonymous") fetchOpts.credentials = "omit";
    else fetchOpts.credentials = "same-origin";
    return fetchOpts;
  }
  function processPreload(link) {
    if (link.ep)
      return;
    link.ep = true;
    const fetchOpts = getFetchOpts(link);
    fetch(link.href, fetchOpts);
  }
})();
function print(method, ...args) {
  if (typeof args[0] === "string") {
    const message = args.shift();
    method(`[wxt] ${message}`, ...args);
  } else {
    method("[wxt]", ...args);
  }
}
var logger = {
  debug: (...args) => print(console.debug, ...args),
  log: (...args) => print(console.log, ...args),
  warn: (...args) => print(console.warn, ...args),
  error: (...args) => print(console.error, ...args)
};
function setupWebSocket(onMessage) {
  const serverUrl = `${"ws:"}//${"localhost"}:${3e3}`;
  logger.debug("Connecting to dev server @", serverUrl);
  const ws = new WebSocket(serverUrl, "vite-hmr");
  ws.addEventListener("open", () => {
    logger.debug("Connected to dev server");
  });
  ws.addEventListener("close", () => {
    logger.debug("Disconnected from dev server");
  });
  ws.addEventListener("error", (event) => {
    logger.error("Failed to connect to dev server", event);
  });
  ws.addEventListener("message", (e) => {
    var _a, _b;
    try {
      const message = JSON.parse(e.data);
      if (message.type === "custom" && ((_b = (_a = message.event) == null ? void 0 : _a.startsWith) == null ? void 0 : _b.call(_a, "wxt:"))) {
        onMessage == null ? void 0 : onMessage(message);
      }
    } catch (err) {
      logger.error("Failed to handle message", err);
    }
  });
  return ws;
}
{
  try {
    setupWebSocket((message) => {
      if (message.event === "wxt:reload-page") {
        if (message.data === location.pathname.substring(1)) {
          location.reload();
        }
      }
    });
  } catch (err) {
    logger.error("Failed to setup web socket connection with dev server", err);
  }
}
//# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoicG9wdXAtRGt1blNxNkkuanMiLCJzb3VyY2VzIjpbIi4uLy4uLy4uLy4uLy4uL25vZGVfbW9kdWxlcy8ucG5wbS93eHRAMC4xNy4xMl9AdHlwZXMrbm9kZUAyNS4yLjFfcm9sbHVwQDQuNTcuMS9ub2RlX21vZHVsZXMvd3h0L2Rpc3QvdmlydHVhbC9yZWxvYWQtaHRtbC5qcyJdLCJzb3VyY2VzQ29udGVudCI6WyIvLyBzcmMvc2FuZGJveC91dGlscy9sb2dnZXIudHNcbmZ1bmN0aW9uIHByaW50KG1ldGhvZCwgLi4uYXJncykge1xuICBpZiAoaW1wb3J0Lm1ldGEuZW52Lk1PREUgPT09IFwicHJvZHVjdGlvblwiKVxuICAgIHJldHVybjtcbiAgaWYgKHR5cGVvZiBhcmdzWzBdID09PSBcInN0cmluZ1wiKSB7XG4gICAgY29uc3QgbWVzc2FnZSA9IGFyZ3Muc2hpZnQoKTtcbiAgICBtZXRob2QoYFt3eHRdICR7bWVzc2FnZX1gLCAuLi5hcmdzKTtcbiAgfSBlbHNlIHtcbiAgICBtZXRob2QoXCJbd3h0XVwiLCAuLi5hcmdzKTtcbiAgfVxufVxudmFyIGxvZ2dlciA9IHtcbiAgZGVidWc6ICguLi5hcmdzKSA9PiBwcmludChjb25zb2xlLmRlYnVnLCAuLi5hcmdzKSxcbiAgbG9nOiAoLi4uYXJncykgPT4gcHJpbnQoY29uc29sZS5sb2csIC4uLmFyZ3MpLFxuICB3YXJuOiAoLi4uYXJncykgPT4gcHJpbnQoY29uc29sZS53YXJuLCAuLi5hcmdzKSxcbiAgZXJyb3I6ICguLi5hcmdzKSA9PiBwcmludChjb25zb2xlLmVycm9yLCAuLi5hcmdzKVxufTtcblxuLy8gc3JjL3ZpcnR1YWwvdXRpbHMvc2V0dXAtd2ViLXNvY2tldC50c1xuZnVuY3Rpb24gc2V0dXBXZWJTb2NrZXQob25NZXNzYWdlKSB7XG4gIGNvbnN0IHNlcnZlclVybCA9IGAke19fREVWX1NFUlZFUl9QUk9UT0NPTF9ffS8vJHtfX0RFVl9TRVJWRVJfSE9TVE5BTUVfX306JHtfX0RFVl9TRVJWRVJfUE9SVF9ffWA7XG4gIGxvZ2dlci5kZWJ1ZyhcIkNvbm5lY3RpbmcgdG8gZGV2IHNlcnZlciBAXCIsIHNlcnZlclVybCk7XG4gIGNvbnN0IHdzID0gbmV3IFdlYlNvY2tldChzZXJ2ZXJVcmwsIFwidml0ZS1obXJcIik7XG4gIHdzLmFkZEV2ZW50TGlzdGVuZXIoXCJvcGVuXCIsICgpID0+IHtcbiAgICBsb2dnZXIuZGVidWcoXCJDb25uZWN0ZWQgdG8gZGV2IHNlcnZlclwiKTtcbiAgfSk7XG4gIHdzLmFkZEV2ZW50TGlzdGVuZXIoXCJjbG9zZVwiLCAoKSA9PiB7XG4gICAgbG9nZ2VyLmRlYnVnKFwiRGlzY29ubmVjdGVkIGZyb20gZGV2IHNlcnZlclwiKTtcbiAgfSk7XG4gIHdzLmFkZEV2ZW50TGlzdGVuZXIoXCJlcnJvclwiLCAoZXZlbnQpID0+IHtcbiAgICBsb2dnZXIuZXJyb3IoXCJGYWlsZWQgdG8gY29ubmVjdCB0byBkZXYgc2VydmVyXCIsIGV2ZW50KTtcbiAgfSk7XG4gIHdzLmFkZEV2ZW50TGlzdGVuZXIoXCJtZXNzYWdlXCIsIChlKSA9PiB7XG4gICAgdHJ5IHtcbiAgICAgIGNvbnN0IG1lc3NhZ2UgPSBKU09OLnBhcnNlKGUuZGF0YSk7XG4gICAgICBpZiAobWVzc2FnZS50eXBlID09PSBcImN1c3RvbVwiICYmIG1lc3NhZ2UuZXZlbnQ/LnN0YXJ0c1dpdGg/LihcInd4dDpcIikpIHtcbiAgICAgICAgb25NZXNzYWdlPy4obWVzc2FnZSk7XG4gICAgICB9XG4gICAgfSBjYXRjaCAoZXJyKSB7XG4gICAgICBsb2dnZXIuZXJyb3IoXCJGYWlsZWQgdG8gaGFuZGxlIG1lc3NhZ2VcIiwgZXJyKTtcbiAgICB9XG4gIH0pO1xuICByZXR1cm4gd3M7XG59XG5cbi8vIHNyYy92aXJ0dWFsL3JlbG9hZC1odG1sLnRzXG5pZiAoaW1wb3J0Lm1ldGEuZW52LkNPTU1BTkQgPT09IFwic2VydmVcIikge1xuICB0cnkge1xuICAgIHNldHVwV2ViU29ja2V0KChtZXNzYWdlKSA9PiB7XG4gICAgICBpZiAobWVzc2FnZS5ldmVudCA9PT0gXCJ3eHQ6cmVsb2FkLXBhZ2VcIikge1xuICAgICAgICBpZiAobWVzc2FnZS5kYXRhID09PSBsb2NhdGlvbi5wYXRobmFtZS5zdWJzdHJpbmcoMSkpIHtcbiAgICAgICAgICBsb2NhdGlvbi5yZWxvYWQoKTtcbiAgICAgICAgfVxuICAgICAgfVxuICAgIH0pO1xuICB9IGNhdGNoIChlcnIpIHtcbiAgICBsb2dnZXIuZXJyb3IoXCJGYWlsZWQgdG8gc2V0dXAgd2ViIHNvY2tldCBjb25uZWN0aW9uIHdpdGggZGV2IHNlcnZlclwiLCBlcnIpO1xuICB9XG59XG4iXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6Ijs7Ozs7Ozs7Ozs7Ozs7Ozs7Ozs7Ozs7Ozs7Ozs7Ozs7Ozs7OztBQUNBLFNBQVMsTUFBTSxXQUFXLE1BQU07QUFHOUIsTUFBSSxPQUFPLEtBQUssQ0FBQyxNQUFNLFVBQVU7QUFDL0IsVUFBTSxVQUFVLEtBQUssTUFBQTtBQUNyQixXQUFPLFNBQVMsT0FBTyxJQUFJLEdBQUcsSUFBSTtBQUFBLEVBQ3BDLE9BQU87QUFDTCxXQUFPLFNBQVMsR0FBRyxJQUFJO0FBQUEsRUFDekI7QUFDRjtBQUNBLElBQUksU0FBUztBQUFBLEVBQ1gsT0FBTyxJQUFJLFNBQVMsTUFBTSxRQUFRLE9BQU8sR0FBRyxJQUFJO0FBQUEsRUFDaEQsS0FBSyxJQUFJLFNBQVMsTUFBTSxRQUFRLEtBQUssR0FBRyxJQUFJO0FBQUEsRUFDNUMsTUFBTSxJQUFJLFNBQVMsTUFBTSxRQUFRLE1BQU0sR0FBRyxJQUFJO0FBQUEsRUFDOUMsT0FBTyxJQUFJLFNBQVMsTUFBTSxRQUFRLE9BQU8sR0FBRyxJQUFJO0FBQ2xEO0FBR0EsU0FBUyxlQUFlLFdBQVc7QUFDakMsUUFBTSxZQUFZLEdBQUcsS0FBdUIsS0FBSyxXQUF1QixJQUFJLEdBQW1CO0FBQy9GLFNBQU8sTUFBTSw4QkFBOEIsU0FBUztBQUNwRCxRQUFNLEtBQUssSUFBSSxVQUFVLFdBQVcsVUFBVTtBQUM5QyxLQUFHLGlCQUFpQixRQUFRLE1BQU07QUFDaEMsV0FBTyxNQUFNLHlCQUF5QjtBQUFBLEVBQ3hDLENBQUM7QUFDRCxLQUFHLGlCQUFpQixTQUFTLE1BQU07QUFDakMsV0FBTyxNQUFNLDhCQUE4QjtBQUFBLEVBQzdDLENBQUM7QUFDRCxLQUFHLGlCQUFpQixTQUFTLENBQUMsVUFBVTtBQUN0QyxXQUFPLE1BQU0sbUNBQW1DLEtBQUs7QUFBQSxFQUN2RCxDQUFDO0FBQ0QsS0FBRyxpQkFBaUIsV0FBVyxDQUFDLE1BQU07O0FBQ3BDLFFBQUk7QUFDRixZQUFNLFVBQVUsS0FBSyxNQUFNLEVBQUUsSUFBSTtBQUNqQyxVQUFJLFFBQVEsU0FBUyxjQUFZLG1CQUFRLFVBQVIsbUJBQWUsZUFBZiw0QkFBNEIsVUFBUztBQUNwRSwrQ0FBWTtBQUFBLE1BQ2Q7QUFBQSxJQUNGLFNBQVMsS0FBSztBQUNaLGFBQU8sTUFBTSw0QkFBNEIsR0FBRztBQUFBLElBQzlDO0FBQUEsRUFDRixDQUFDO0FBQ0QsU0FBTztBQUNUO0FBR3lDO0FBQ3ZDLE1BQUk7QUFDRixtQkFBZSxDQUFDLFlBQVk7QUFDMUIsVUFBSSxRQUFRLFVBQVUsbUJBQW1CO0FBQ3ZDLFlBQUksUUFBUSxTQUFTLFNBQVMsU0FBUyxVQUFVLENBQUMsR0FBRztBQUNuRCxtQkFBUyxPQUFBO0FBQUEsUUFDWDtBQUFBLE1BQ0Y7QUFBQSxJQUNGLENBQUM7QUFBQSxFQUNILFNBQVMsS0FBSztBQUNaLFdBQU8sTUFBTSx5REFBeUQsR0FBRztBQUFBLEVBQzNFO0FBQ0Y7IiwieF9nb29nbGVfaWdub3JlTGlzdCI6WzBdfQ==
