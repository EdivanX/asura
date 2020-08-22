package commands

import (
	"asura/src/database"
	"asura/src/handler"
	"asura/src/utils"
	"context"
	"fmt"
	"github.com/andersfylling/disgord"
	"strconv"
	"strings"
	"time"
)

var cGuilds []*disgord.Guild
var lastUpdated time.Time = time.Now()

func init() {

	handler.Register(handler.Command{
		Aliases:   []string{"userinfo", "usuario", "uinfo"},
		Run:       runUserinfo,
		Available: true,
		Cooldown:  4,
		Usage:     "j!userinfo <usuario.",
		Help:      "Veja as informaçoes de um usuario",
	})
}

func updateGuilds(session disgord.Session){
	guilds,err := session.GetGuilds(context.Background(), &disgord.GetCurrentUserGuildsParams{})
	if err == nil{
		cGuilds = guilds
		lastUpdated = time.Now()
	}
}
func runUserinfo(session disgord.Session, msg *disgord.Message, args []string) {
	user := utils.GetUser(msg, args, session)
	var userinfo database.User
	var private bool
	avatar, _ := user.AvatarURL(512, false)
	authorAvatar, _ := msg.Author.AvatarURL(512, false)
	id := strconv.FormatUint(uint64(user.ID), 10)
	database.Database.NewRef("users/"+id).Get(context.Background(), &userinfo)
	database.Database.NewRef("private/"+id).Get(context.Background(), &private)
	oldAvatars := ""
	oldUsernames := ""
	filteredGuilds := ""
	count := 0
	if len(cGuilds) == 0 {
		updateGuilds(session)
	}
	if time.Since(lastUpdated).Seconds()/60 > 30 {
		go updateGuilds(session)
	}
	date := ((uint64(user.ID) >> 22) + 1420070400000) / 1000
	for i, guild := range cGuilds {
		_, is := session.GetMember(context.Background(),guild.ID,user.ID)
		if count >= 12 {
			break
		}
		if is == nil {
			filteredGuilds += guild.Name
			count++
			if i != len(cGuilds) {
				filteredGuilds += "** | **"
			}

		}
	}
	if len(userinfo.Usernames) > 0 {
		if len(userinfo.Usernames) > 12 {
			oldUsernames = strings.Join(userinfo.Usernames[:12], "** | **")
		} else {
			oldUsernames = strings.Join(userinfo.Usernames, "** | **")
		}
	} else {
		oldUsernames = "Nenhum nome antigo registrado"
	}
	if len(userinfo.Avatars) > 0 && !private {
		avats := userinfo.Avatars
		if len(avats) > 12 {
			avats = avats[:12]
		}
		for i, avatar := range avats {
			oldAvatars += fmt.Sprintf("[**Link**](%s)", avatar)
			if i != len(avats) {
				oldAvatars += "** | **"
			}
		}
	} else if private {
		oldAvatars = "O historico desse usuario é privado, use j!private para deixar publico"
	} else {
		oldAvatars = "Nenhum avatar antigo registrado"
	}
	msg.Reply(context.Background(), session, &disgord.CreateMessageParams{
		Embed: &disgord.Embed{
			Color: 65535,
			Title: fmt.Sprintf("%s(%s)", user.Username, user.ID),
			Thumbnail: &disgord.EmbedThumbnail{
				URL: avatar,
			},
			Description: fmt.Sprintf("Conta criada a **%d** Dias", int((uint64(time.Now().Unix())-date)/60/60/24)),
			Footer: &disgord.EmbedFooter{
				Text:    msg.Author.Username,
				IconURL: authorAvatar,
			},
			Fields: []*disgord.EmbedField{
				&disgord.EmbedField{
					Name:  "Nomes Antigos",
					Value: oldUsernames,
				},
				&disgord.EmbedField{
					Name:  "Avatares Antigos",
					Value: oldAvatars,
				},
				&disgord.EmbedField{
					Name:  fmt.Sprintf("Servidores compartilhados (%d)", count),
					Value: filteredGuilds,
				},
				&disgord.EmbedField{
					Name:  "Mais informaçoes",
					Value: fmt.Sprintf("[**Clique aqui**](https://asura-site.glitch.me/user/%s)", id),
				},
			},
		}})
}
