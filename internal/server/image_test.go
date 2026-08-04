package server

import "testing"

func TestExtractFirstImage(t *testing.T) {
	cases := []struct {
		name   string
		cooked string
		want   string
	}{
		{
			name:   "обычная картинка",
			cooked: `<p>Текст</p><p><img src="https://gig.ovh/uploads/default/original/1X/abc.png" alt="screenshot" width="690"></p>`,
			want:   "https://gig.ovh/uploads/default/original/1X/abc.png",
		},
		{
			name:   "смайл пропускается",
			cooked: `<p>Привет <img src="https://gig.ovh/images/emoji/twitter/wave.png" class="emoji" alt=":wave:"> всем</p><img src="https://gig.ovh/uploads/pic.jpg">`,
			want:   "https://gig.ovh/uploads/pic.jpg",
		},
		{
			name:   "иконка сайта из onebox пропускается",
			cooked: `<aside class="onebox"><img src="https://example.com/favicon.ico" class="site-icon"></aside>`,
			want:   "",
		},
		{
			name:   "protocol-relative URL дополняется",
			cooked: `<img src="//cdn.example.com/pic.png">`,
			want:   "https://cdn.example.com/pic.png",
		},
		{
			name:   "пост без картинок",
			cooked: `<p>Просто текст</p>`,
			want:   "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := extractFirstImage(tc.cooked); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
