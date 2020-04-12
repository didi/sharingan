
function handleSelect(key, keyPath, activeIndex) {
    if(key == activeIndex) {
        return
    }
    switch(key) {
        case "1":
            window.open(redirect("/"), "_blank")
            break
        case "3":
            window.open(redirect("/mock"), "_blank")
            break
        case "80":
            window.open(redirect("/manual"), "_blank")
            break
        case "99-1":
            upgrade()
            break
    }
}

function upgrade() {
    axios({
        method: "get",
        url: "/upgrade",
    }).then(function (response) {
        // TODO: how to retry and refresh
    }).catch(function (error) {
        console.error(error)
    })
}

function redirect(url) {
    var host = "8998"
    return document.location.href.substring(0, document.location.href.indexOf(":"+host+"/")+5) + url;
}

export { handleSelect }
