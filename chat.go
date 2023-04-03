package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"sync"
	"time"
)

const (
	MainStatus     = ChatStatus(iota)
	TrainStatus    = ChatStatus(iota)
	SettingsStatus = ChatStatus(iota)
	nilStatus      = ChatStatus(iota)
	showStatus     = ChatStatus(iota)
)

type ChatStatus uint8

type Chat struct {
	id         int64
	channel    chan tgbotapi.Update
	timeStart  time.Time
	timeFinish time.Time
}

var startKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(StartTrainText)),
	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(showText)),
	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(TrainSettingsText)),
)

var SettingsKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(AddExeciseText)),
	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(RemoveExerciseText)),
	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(BackText)),
)

var confirmationKayboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(YesText)),
	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(NoText)),
)

var attributeKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(ApproachesCountText),
		tgbotapi.NewKeyboardButton(RepeatCountText)),
	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(TimeText),
		tgbotapi.NewKeyboardButton(DistanceText), tgbotapi.NewKeyboardButton(WeightText)),
	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(SaveText)),
)

var whatToShowKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(showByDateText),
		tgbotapi.NewKeyboardButton(showDatesText)),
	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(showExerciseStatisticsText)),
	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(BackText)),
)

var commentSkipKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(SkipText)),
)

var todayKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(todayText)),
)

var timeAttributeKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(TimeEndText)),
	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(InputTimeByUserText)),
)

const (
	helloText = "Привет, я Олег Жёсткожимов, помогаю братанам с сохранением и анализом своих результатов в качалочке!\n" +
		"Сначала зайди в настройки и добавь упражнения, разберёшься там, я буду тебе помогать\n" +
		"Ну а потом можешь и потренироваться, нажав на кнопку \"Начать тренировку\", там я тебе тоже помогу, выберешь упражнение" +
		", запишешь свои результаты. Чуть не забыл, важно отметить, что помимо того, что нужно быть аккуратнее при работе" +
		"с весами, ещё если ты меняешь вес, то записывай это как начало нового упражнения, просто мне так проще ориентироваться" +
		" в данных, заранее благодарю! Ну удачных тренировок тебе, дружище, если что, ты знаешь где меня искать, " +
		"я всегда буду тут!"

	StartTrainText      = "Начало тренировки"
	TrainSettingsText   = "Настройки"
	AddExeciseText      = "Добавить упражнение"
	BackText            = "Назад"
	YesText             = "Да"
	NoText              = "Нет"
	RemoveExerciseText  = "Удалить упражнение"
	ApproachesCountText = "Подходы"
	RepeatCountText     = "Повторения"
	TimeText            = "Время"
	DistanceText        = "Дистанция"
	WeightText          = "Вес"
	SaveText            = "Сохранить"

	showText                   = "Результаты"
	showByDateText             = "За период времени\n(пока за два года больше нихуя не сделали да...)"
	showDatesText              = "Когда я тренируюсь"
	showExerciseStatisticsText = "Статистика упражнений"

	RequestExName = "Введи название упражнение, мужик"

	SkipText = "Пропустить"

	todayText = "Сегодня"

	TimeEndText         = "Закончить"
	InputTimeByUserText = "Ввести время вручную"
)

/*showCmd.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
[]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(StartTrainText, "fff")})*/

func wrongInput(id int64) {
	BotAPI.Send(tgbotapi.NewMessage(id, "Не то ввёл, братишка"))
}

func ShowMainScreen(id int64) ChatStatus {
	showCmd := tgbotapi.NewMessage(id, "Работаем, братишка")
	showCmd.ReplyMarkup = startKeyboard
	BotAPI.Request(showCmd)
	return MainStatus
}

func ShowSettingsScreen(id int64, mes string) ChatStatus {
	showCmd := tgbotapi.NewMessage(id, mes)
	showCmd.ReplyMarkup = SettingsKeyboard
	BotAPI.Request(showCmd)
	return SettingsStatus
}

func ShowStartTrainScreen(id int64, db *sql.DB) (ChatStatus, []string) {
	showCmd := tgbotapi.NewMessage(id, "Поехали, братишка")
	Keyboards := [][]tgbotapi.KeyboardButton{}
	rows, err := db.Query(`select "exercise" from "exercises" where chatid = $1`, id)
	if err != nil {
		log.Println(err.Error())
	}
	var str string
	var execs []string
	for rows.Next() {
		rows.Scan(&str)
		log.Println(str)
		Keyboards = append(Keyboards, []tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton(str)})
		execs = append(execs, str)
	}
	Keyboards = append(Keyboards, []tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton(BackText)})
	showCmd.ReplyMarkup = tgbotapi.NewReplyKeyboard(Keyboards...)
	BotAPI.Request(showCmd)
	return TrainStatus, execs
}

func AddExersiceStage2(db *sql.DB, id int64, newExecise string) (ChatStatus, int32) {
	var execId int32
	row := db.QueryRow(`insert into "exercises"("chatid", "exercise") values ($1, $2) returning "id"`, id, newExecise)
	err := row.Scan(&execId)
	if err != nil {
		log.Println(err.Error())
	}
	BotAPI.Send(tgbotapi.NewMessage(id, "Упражнение "+newExecise+" добавлено"))
	showCmd := tgbotapi.NewMessage(id, "Что нам для этого нужно, дружище?")
	showCmd.ReplyMarkup = attributeKeyboard
	BotAPI.Request(showCmd)
	return SettingsStatus, execId
}

func AddExersiceDenied(id int64) ChatStatus {
	BotAPI.Send(tgbotapi.NewMessage(id, "Ладно, это нам и не надо, правильно, братан!"))
	showCmd := tgbotapi.NewMessage(id, "Чё делаем теперь?")
	showCmd.ReplyMarkup = SettingsKeyboard
	BotAPI.Request(showCmd)
	return SettingsStatus
}

func InputText(id int64, channel chan tgbotapi.Update, disc string) (string, error) {
	BotAPI.Send(tgbotapi.NewMessage(id, disc))
	message := <-channel
	if message.Message == nil {
		return "", errors.New("Не то ввел, братишка")
	}
	return message.Message.Text, nil
}

func ShowAddExerciseScreen(id int64, channel chan tgbotapi.Update) (string, string) {
	newExecise, err := InputText(id, channel, RequestExName)
	if newExecise == BackText {
		AddExersiceDenied(id)
		return "", ""
	}
	if err != nil {
		BotAPI.Send(tgbotapi.NewMessage(id, err.Error()))
	}
	showCmd := tgbotapi.NewMessage(id, "Значит добавляем упражнение "+
		newExecise+", братишка?")
	showCmd.ReplyMarkup = confirmationKayboard
	BotAPI.Request(showCmd)
	temp, err := InputText(id, channel, "")
	if err != nil {
		wrongInput(id)
	}
	return temp, newExecise
}

func MillisecondsToTime(milliseconds int64) time.Time {
	return time.Unix(0, milliseconds*int64(time.Millisecond))
}

func TrainProcess(id int64, channal chan tgbotapi.Update, execs []string, db *sql.DB, trainID int64) ChatStatus {
	for {
		var attributes []string
		var exec string
		var execId int64
		var l = 0
		str, err := InputText(id, channal, "")
		if err != nil {
			log.Println(err.Error())
			wrongInput(id)
		}
		if str == BackText {
			showCmd := tgbotapi.NewMessage(id, "Как опишешь тренировочку, братишка?")
			showCmd.ReplyMarkup = commentSkipKeyboard
			BotAPI.Request(showCmd)
			comment, err := InputText(id, channal, "")
			if err != nil {
				log.Println(err.Error())
			}
			if comment != SkipText {
				db.Exec(`update "state" set "comment" = $1 where "id" = $2`, comment, trainID)
			}
			db.Exec(`update "state" set "endtime" = $1 where "id"=$2`,
				time.Now(), trainID)
			return ShowMainScreen(id)
		}

		for {
			if l >= len(execs) {
				BotAPI.Send(tgbotapi.NewMessage(id, "Такого упражнения нет, корефан :("))
				break
			}
			if str == execs[l] {
				row := db.QueryRow(`select "id" from "exercises" where "exercise" = $1`, str)
				err = row.Scan(&execId)
				rows, err := db.Query(`select "name" from "attributes" where "exercise" = $1`, execId)
				if err != nil {
					log.Println(err.Error())
				}
				for rows.Next() {
					rows.Scan(&exec)
					attributes = append(attributes, exec)
				}
				timeSearchIndex := 0
				timeExist := false
				for {
					if attributes[timeSearchIndex] == TimeText {
						timeExist = true
						break
					}
					timeSearchIndex++
				}
				if timeExist {
					showCmd := tgbotapi.NewMessage(id, "Отсчёт времени пошёл, братэлло!")
					showCmd.ReplyMarkup = timeAttributeKeyboard
					BotAPI.Request(showCmd)
					timeStart := time.Now()
					str, err := InputText(id, channal, "")
					if err != nil {
						log.Println(err.Error())
					}
					switch str {
					case TimeEndText:
						timeEnd := time.Now()
						timeOfExec := timeEnd.Sub(timeStart)
						timeOfExec = timeOfExec.Truncate(time.Nanosecond * 10)
						milliTime := int64(timeOfExec.Milliseconds())
						BotAPI.Send(tgbotapi.NewMessage(id, "Закончили, отдыхаем!\nВремя составило "+
							MillisecondsToTime(milliTime).Format("04:05.06")))
						db.Exec(`insert into "results"("trainid", "execid", "attribute", "value") values ($1, $2, $3, $4)`,
							trainID, execId, "Время", milliTime)
					case InputTimeByUserText:
						timeStr, err := InputText(id, channal, "Введи время, бро\nШаблон: 01:02:03, 02:03.004")
						if err != nil {
							log.Println(err.Error())
						}
						timeT, err := time.Parse("04:05.06", timeStr)
						if err != nil {
							log.Println(err.Error())
						}
						log.Println(timeT)
					}
				}
				i := 0
				for {
					if i >= len(attributes) {
						break
					}

					if attributes[i] != TimeText {
						value, err := InputText(id, channal, attributes[i]+"?")
						if err != nil {
							log.Println(err.Error())
						}
						valueInt, err := strconv.Atoi(value)
						if err != nil {
							log.Println(err.Error())
						}
						_, err = db.Exec(`insert into "results"("trainid", "execid", "attribute", "value") values ($1, $2, $3, $4)`,
							trainID, execId, attributes[i], valueInt)
						if err != nil {
							log.Println("ОПАЧКИ " + err.Error())
						}
						if err != nil {
							log.Println(err.Error())
						}
					}
					i++
				}
				ShowStartTrainScreen(id, db)
				break
			} else {
				l++
			}
		}
	}
}

func getExersises(db *sql.DB, id int64, disc string) {
	showCmd := tgbotapi.NewMessage(id, disc)
	Keyboards := [][]tgbotapi.KeyboardButton{}
	rows, err := db.Query(`select "exercise" from "exercises" where chatid = $1`, id)
	if err != nil {
		log.Println(err.Error())
	}
	var str string
	for rows.Next() {
		rows.Scan(&str)
		log.Println(str)
		Keyboards = append(Keyboards, []tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton(str)})
	}
	Keyboards = append(Keyboards, []tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton(BackText)})
	showCmd.ReplyMarkup = tgbotapi.NewReplyKeyboard(Keyboards...)
	BotAPI.Request(showCmd)
}

func showRemoveExeciseScreen(id int64, db *sql.DB, channel chan tgbotapi.Update) {
	getExersises(db, id, "Что убираем, дружище?")
	name, err := InputText(id, channel, "")
	if err != nil {
		wrongInput(id)
	}
	var myID int32
	row := db.QueryRow(`select "id" from "exercises" where "exercise" = $1 and chatid = $2`, name, id)
	err = row.Scan(&myID)
	_, err = db.Exec(`delete from "attributes" where "exercise" = $1`, myID)
	if err != nil {
		log.Println(err.Error())
	}
	_, err = db.Exec(`delete  from "results" where "execid" = $1`, myID)
	if err != nil {
		log.Println(err.Error())
	}
	_, err = db.Exec(`delete from "exercises" where "exercise" = $1 and chatid = $2`, name, id)
	if err != nil {
		log.Println(err.Error())
	}
	showCmd := tgbotapi.NewMessage(id, "Удалили, давай теперь назад!")
	showCmd.ReplyMarkup = SettingsKeyboard
	BotAPI.Request(showCmd)
}

func showAddAtributesScreen(id int64, channel chan tgbotapi.Update, db *sql.DB, execId int32) ChatStatus {
	for {
		str, err := InputText(id, channel, "")
		if str == SaveText {
			return ShowSettingsScreen(id, "Нармуль, готово!")
		}
		if err != nil {
			wrongInput(id)
		}
		db.Exec(`insert into "attributes"(exercise, name) values($1, $2)`, execId, str)
		BotAPI.Send(tgbotapi.NewMessage(id, "Понял"))
	}
}
func showMe() {

}

func (chat Chat) routine(chats map[int64]Chat, mainMutex *sync.Mutex, db *sql.DB) {
	lastMassageTime := time.After(time.Hour * 10)
	status := nilStatus
	var execId int32
	var execs []string
	var trainId int64
	for {
		select {
		case message := <-chat.channel:
			lastMassageTime = time.After(time.Hour * 10)
			log.Println("MainStatus.Update")
			if message.Message != nil {
				switch status {
				case nilStatus:
					switch message.Message.Text {
					case "/start":
						BotAPI.Send(tgbotapi.NewMessage(chat.id, helloText))
						status = ShowMainScreen(chat.id)
					}
				case MainStatus:
					switch message.Message.Text {
					case "/start":
						status = ShowMainScreen(chat.id)
					case TrainSettingsText:
						status = ShowSettingsScreen(chat.id, "Ща настроимся, корефан!")
					case StartTrainText:
						row := db.QueryRow(`insert into "state"(date, starttime, userid) values($1, $2, $3)
							 returning  "id"`, time.Now(), time.Now(), chat.id)
						row.Scan(&trainId)
						status, execs = ShowStartTrainScreen(chat.id, db)
						status = TrainProcess(chat.id, chat.channel, execs, db, trainId)
					case showText:
						status = showStatus
						showCmd := tgbotapi.NewMessage(chat.id, "Что хочешь узнать, комрад?")
						showCmd.ReplyMarkup = whatToShowKeyboard
						BotAPI.Request(showCmd)
						toShow, err := InputText(chat.id, chat.channel, "")
						if err != nil {
							log.Println(err.Error())
						}
						switch toShow {
						case showByDateText:
						//dateTxt, err := InputText(chat.id, chat.channel,
						//	"Введи дату с которой начинаем отсчёт, бро\n(дд.мм.гггг)")
						//showCmd = tgbotapi.NewMessage(chat.id,
						//	"Введи дату с которой до которой смотрим, бро\n(дд.мм.гггг)")
						//showCmd.ReplyMarkup = todayText
						//BotAPI.Request(showCmd)
						//dateEndTxt, err := InputText(chat.id, chat.channel, "")
						//if dateEndTxt == todayText {
						//	dateEnd := time.Now()
						//}
						case showExerciseStatisticsText:
							var IDexec int32
							var attributeNames []string
							var values []string
							var IDtrain []string
							var dates []string
							attributeValue := 0
							var temp string
							getExersises(db, chat.id, "Какое упражнение нас итересует, дружище?")
							exercise, err := InputText(chat.id, chat.channel, "")
							if err != nil {
								log.Println(err.Error())
							}

							row := db.QueryRow(`select "id" from "exercises" where "exercise" = $1 and chatid = $2`,
								exercise, chat.id)
							row.Scan(&IDexec)
							rows, err := db.Query(`select "name" from "attributes" where "exercise" = $1`, IDexec)
							for rows.Next() {
								attributeValue++
							}
							rows, err = db.Query(`select "attribute" from "results" where "execid" = $1`, IDexec)
							if err != nil {
								log.Println(err.Error())
							}
							for rows.Next() {
								rows.Scan(&temp)
								attributeNames = append(attributeNames, temp)
							}
							rows, err = db.Query(`select "value" from "results" where execid = $1`, IDexec)
							if err != nil {
								log.Println(err.Error())
							}
							for rows.Next() {
								rows.Scan(&temp)
								values = append(values, temp)
							}
							rows, err = db.Query(`select "trainid" from "results" where "execid" = $1`, IDexec)
							if err != nil {
								log.Println(err.Error())
							}
							for rows.Next() {
								rows.Scan(&temp)
								IDtrain = append(IDtrain, temp)
							}
							i := 0
							var date time.Time
							for {
								if i >= len(IDtrain) {
									break
								}
								rows, err = db.Query(`select "date" from "state" where "id" = $1`, IDtrain[i])
								if err != nil {
									log.Println(err.Error())
								}
								for rows.Next() {
									rows.Scan(&date)
									dates = append(dates, date.Format("02.01.06"))
								}
								i++
							}
							result := bytes.NewBufferString("Вот что я помню по упражнению " + exercise + ":\n")
							iterator := 0
							for {
								if iterator >= len(values) {
									break
								}
								fmt.Fprint(result, "\n"+dates[iterator]+":\n")
								j := 0
								for {
									if j >= attributeValue {
										break
									}
									if attributeNames[iterator+j] == TimeText {
										var hours int
										var min int
										var sec int
										var msec int
										timeint, err := strconv.Atoi(values[iterator+j])
										if err != nil {
											log.Println(err.Error())
										}
										msec = timeint % 1000
										sec = timeint / 1000
										min = sec / 60
										sec = sec % 60
										hours = min / 60
										min = min / 60
										fmt.Fprint(result, "   "+attributeNames[iterator+j]+" - "+
											strconv.Itoa(hours)+":"+strconv.Itoa(min)+":"+strconv.Itoa(sec)+"."+strconv.Itoa(msec)+"\n")
									} else {
										fmt.Fprint(result, "   "+attributeNames[iterator+j]+" - "+values[iterator+j]+"\n")
									}
									j++
								}
								iterator += j
							}
							BotAPI.Send(tgbotapi.NewMessage(chat.id, result.String()))
						case showDatesText:
							var temp time.Time
							var dates []string
							rows, err := db.Query(`select "date" from "state" where "userid" = $1`, chat.id)
							if err != nil {
								log.Println(err.Error())
							}
							for rows.Next() {
								rows.Scan(&temp)
								dates = append(dates, temp.Format("02.01.06"))
							}
							result := bytes.NewBufferString(dates[0] + "\n")
							i := 1
							for {
								if i > len(dates) {
									break
								}
								fmt.Fprint(result, dates[i]+"\n")
								i++
							}
							BotAPI.Send(tgbotapi.NewMessage(chat.id, result.String()))
						case BackText:
							status = ShowMainScreen(chat.id)
						}
						status = ShowMainScreen(chat.id)
					}
				case SettingsStatus:
					switch message.Message.Text {
					case AddExeciseText:
						temp, newExecise := ShowAddExerciseScreen(chat.id, chat.channel)
						switch temp {
						case YesText:
							status, execId = AddExersiceStage2(db, chat.id, newExecise)
							showAddAtributesScreen(chat.id, chat.channel, db, execId)
						case NoText:
							AddExersiceDenied(chat.id)
						}
					case RemoveExerciseText:
						showRemoveExeciseScreen(chat.id, db, chat.channel)
					case BackText:
						status = ShowMainScreen(chat.id)
					default:
						wrongInput(chat.id)
					}
				case TrainStatus:
					switch message.Message.Text {
					case BackText:
						status = ShowMainScreen(chat.id)
					default:
						rows, err := db.Query(`select "exercise" from "exercises" where
                                      exercise = $1 and chatid = $2`, message.Message.Text, chat.id)
						if err != nil {
							log.Println(err.Error())
						}
						log.Println(rows)
					}
				}
			}

		case <-lastMassageTime:
			log.Println("time out")
			mainMutex.Lock()
			log.Println("Chat " + strconv.FormatInt(chat.id, 10) + " deleted")
			log.Println(chats)
			delete(chats, chat.id)
			mainMutex.Unlock()
		}
	}
}
