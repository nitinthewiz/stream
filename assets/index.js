if (localStorage.theme === 'dark' || (!('theme' in localStorage) && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
    document.documentElement.classList.add('dark')
} else {
    document.documentElement.classList.remove('dark')
}

var inMemoryPosts;
var currentlyEditingPostID = 0;

fetch('/posts')
    .then((response) => response.json())
        .then((data) => {
            localStorage.posts = JSON.stringify(data);
            inMemoryPosts = data;
            for (post of data) {
                if (post.contentHTML && post.contentHTML !== '') {
                    document.getElementById(post.ID).innerHTML = post.contentHTML;
                }
                if (moment()) {
                    var date_data = document.getElementById('date-'+post.ID).innerHTML;
                    var moment_date = moment(date_data)
                    if (moment_date.isValid()){
                        document.getElementById('date-'+post.ID).innerHTML = moment_date.fromNow();
                    }
                }
            }
        });



document.querySelector('.submitButton')?.addEventListener('click', handleSubmit);
document.querySelector('.toggleDarkMode')?.addEventListener('click', handleDarkModeToggle);
document.querySelector('.postsCol').addEventListener('click', event => {
        if (event.target.matches('.editButton')) {
            handleEdit(event.target);
        }
});

async function handleSubmit() {
    var fetchEndpoint = '/posts';
    var fetchMethod = 'POST';
    if (document.querySelector('.textareaClass').value == ''){
        return;
    }

    var data = {
            "author": "me",
            "content": document.querySelector('.textareaClass').value
    }
    if (currentlyEditingPostID != 0){
        fetchEndpoint = fetchEndpoint + '/' + currentlyEditingPostID.toString();
        fetchMethod = 'PUT';
    }
    const response = await fetch(fetchEndpoint, {
                        method: fetchMethod,
                        headers: {
                            'Content-Type': 'application/json',
                            'Token': localStorage.current_secret
                        },
                        body: JSON.stringify(data)
                    });
    if (!response.ok) {
    }
    const body = await response.text();
    document.querySelector('.textareaClass').value = '';
    currentlyEditingPostID = 0;
    location.reload();
}

async function handleEdit(element){
    var foundEditingPostKey = -1;
    currentlyEditingPostID = parseInt(element.dataset.id);
    if (document.querySelector('.textareaClass').value == ''){
        document.querySelector('.textareaClass').value = '';
    }
    if (inMemoryPosts == null || inMemoryPosts == ""){
        inMemoryPosts = JSON.parse(localStorage.posts);
    }
    for (key in inMemoryPosts) {
        if (inMemoryPosts[key].ID == currentlyEditingPostID){
            foundEditingPostKey = key;
            break;
        }
    }
    if (foundEditingPostKey == -1){
        alert("Could not locate the post you're trying to edit. Maybe refresh the page before trying again.");
    } else {
        post = inMemoryPosts[foundEditingPostKey];
        document.querySelector('.textareaClass').value = post.content;
        window.scrollTo({ top: 0, behavior: 'smooth' });
    }
}

async function handleDarkModeToggle(){
    if (localStorage.theme === 'dark') {
        localStorage.theme = 'light';
    } else {
        localStorage.theme = 'dark';
    }
    location.reload();
}
