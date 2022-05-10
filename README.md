# proxy-router

### feature

1. health check
1. auto select best one
1. extract proxies from file/url with delimiter like `\n`
1. write script to extract proxies from file/url
```javascript
      function extract(doc) {
          var regex = /host=([\d\.]+)&port=(\d+)/g;
          var matched = doc.matchAll(regex);
          var result = [];
          for (const element of matched) {
            var proxy = {"type":"http", "server": `${element[1]}:${element[2]}`};
            result.push(proxy);
          }
          return result;
        }
```
  