package homeassistant

type shortServicesStruct struct {
	XiaomiMiot  XiaomiMiot  `json:"xiaomi_miot"`
	Switch      Switch      `json:"switch"`
	Light       Light       `json:"light"`
	Fan         Fan         `json:"fan"`
	Climate     Climate     `json:"climate"`
	MediaPlayer MediaPlayer `json:"media_player"`
	Select      Select      `json:"select"`
}

// 红外控制空调选择会用到
type Select struct {
	Domain   string `json:"domain"`
	Services struct {
		SelectFirst struct {
			Fields interface{} `json:"fields"`
		} `json:"select_first"`
		SelectLast struct {
			Fields interface{} `json:"fields"`
		} `json:"select_last"`
		SelectNext struct {
			Fields struct {
				Cycle struct {
					Default bool `json:"default"`

					Selector struct {
						Boolean interface{} `json:"boolean"`
					} `json:"selector"`
				} `json:"cycle"`
			} `json:"fields"`
		} `json:"select_next"`
		SelectOption struct {
			Fields struct {
				Option struct {
					Example  string `json:"example"`
					Required bool   `json:"required"`
					Selector struct {
						Text interface{} `json:"text"`
					} `json:"selector"`
				} `json:"option"`
			} `json:"fields"`
		} `json:"select_option"`
		SelectPrevious struct {
			Fields struct {
				Cycle struct {
					Default  bool `json:"default"`
					Selector struct {
						Boolean interface{} `json:"boolean"`
					} `json:"selector"`
				} `json:"cycle"`
			} `json:"fields"`
		} `json:"select_previous"`
	} `json:"services"`
}

type MediaPlayer struct {
	Domain   string `json:"domain"`
	Services struct {
		ClearPlaylist struct {
			Fields struct {
			} `json:"fields"`
		} `json:"clear_playlist"`
		Join struct {
			Fields struct {
				GroupMembers struct {
					Example  string `json:"example"`
					Required bool   `json:"required"`
				} `json:"group_members"`
			} `json:"fields"`
		} `json:"join"`
		MediaNextTrack     struct{} `json:"media_next_track"`
		MediaPause         struct{} `json:"media_pause"`
		MediaPlay          struct{} `json:"media_play"`
		MediaPlayPause     struct{} `json:"media_play_pause"`
		MediaPreviousTrack struct{} `json:"media_previous_track"`
		MediaSeek          struct {
			Fields struct {
				SeekPosition struct {
					Required bool `json:"required"`
					Selector struct {
						Number struct {
							Max  int64   `json:"max"`
							Min  int     `json:"min"`
							Mode string  `json:"mode"`
							Step float64 `json:"step"`
						} `json:"number"`
					} `json:"selector"`
				} `json:"seek_position"`
			} `json:"fields"`
		} `json:"media_seek"`
		MediaStop struct{} `json:"media_stop"`
		PlayMedia struct {
			Fields struct {
				Announce struct {
					Example string `json:"example"`
					Filter  struct {
						SupportedFeatures []int `json:"supported_features"`
					} `json:"filter"`

					Required bool `json:"required"`
					Selector struct {
						Boolean interface{} `json:"boolean"`
					} `json:"selector"`
				} `json:"announce"`
				Enqueue struct {
					Filter struct {
						SupportedFeatures []int `json:"supported_features"`
					} `json:"filter"`

					Required bool `json:"required"`
					Selector struct {
						Select struct {
							Options        []string `json:"options"`
							TranslationKey string   `json:"translation_key"`
						} `json:"select"`
					} `json:"selector"`
				} `json:"enqueue"`
				MediaContentID struct {
					Example string `json:"example"`

					Required bool `json:"required"`
					Selector struct {
						Text interface{} `json:"text"`
					} `json:"selector"`
				} `json:"media_content_id"`
				MediaContentType struct {
					Example string `json:"example"`

					Required bool `json:"required"`
					Selector struct {
						Text interface{} `json:"text"`
					} `json:"selector"`
				} `json:"media_content_type"`
			} `json:"fields"`
			Name string `json:"name"`
		} `json:"play_media"`
		RepeatSet struct {
			Fields struct {
				Repeat struct {
					Required bool `json:"required"`
					Selector struct {
						Select struct {
							Options        []string `json:"options"`
							TranslationKey string   `json:"translation_key"`
						} `json:"select"`
					} `json:"selector"`
				} `json:"repeat"`
			} `json:"fields"`
			Name string `json:"name"`
		} `json:"repeat_set"`
		SelectSoundMode struct {
			Fields struct {
				SoundMode struct {
					Example string `json:"example"`

					Selector struct {
						Text interface{} `json:"text"`
					} `json:"selector"`
				} `json:"sound_mode"`
			} `json:"fields"`
		} `json:"select_sound_mode"`
		SelectSource struct {
			Fields struct {
				Source struct {
					Example  string `json:"example"`
					Required bool   `json:"required"`
					Selector struct {
						Text interface{} `json:"text"`
					} `json:"selector"`
				} `json:"source"`
			} `json:"fields"`
		} `json:"select_source"`
		ShuffleSet struct {
			Fields struct {
				Shuffle struct {
					Required bool `json:"required"`
					Selector struct {
						Boolean interface{} `json:"boolean"`
					} `json:"selector"`
				} `json:"shuffle"`
			} `json:"fields"`
		} `json:"shuffle_set"`
		Toggle     struct{} `json:"toggle"`
		TurnOff    struct{} `json:"turn_off"`
		TurnOn     struct{} `json:"turn_on"`
		Unjoin     struct{} `json:"unjoin"`
		VolumeDown struct{} `json:"volume_down"`
		VolumeMute struct{} `json:"volume_mute"`
		VolumeSet  struct {
			Fields struct {
				VolumeLevel struct {
					Required bool `json:"required"`
					Selector struct {
						Number struct {
							Max  int     `json:"max"`
							Min  int     `json:"min"`
							Step float64 `json:"step"`
						} `json:"number"`
					} `json:"selector"`
				} `json:"volume_level"`
			} `json:"fields"`
		} `json:"volume_set"`
		VolumeUp struct{} `json:"volume_up"`
	} `json:"services"`
}

type Climate struct {
	Domain   string `json:"domain"`
	Services struct {
		SetAuxHeat struct {
			Fields struct {
				AuxHeat struct {
					Required bool `json:"required"`
					Selector struct {
						Boolean interface{} `json:"boolean"`
					} `json:"selector"`
				} `json:"aux_heat"`
			} `json:"fields"`
		} `json:"set_aux_heat"`
		SetFanMode struct {
			Fields struct {
				FanMode struct {
					Example string `json:"example"`

					Required bool `json:"required"`
					Selector struct {
						Text interface{} `json:"text"`
					} `json:"selector"`
				} `json:"fan_mode"`
			} `json:"fields"`
		} `json:"set_fan_mode"`
		SetHumidity struct {
			Fields struct {
				Humidity struct {
					Required bool `json:"required"`
					Selector struct {
						Number struct {
							Max               int    `json:"max"`
							Min               int    `json:"min"`
							UnitOfMeasurement string `json:"unit_of_measurement"`
						} `json:"number"`
					} `json:"selector"`
				} `json:"humidity"`
			} `json:"fields"`
		} `json:"set_humidity"`
		SetHvacMode struct {
			Fields struct {
				HvacMode struct {
					Selector struct {
						Select struct {
							Options        []string `json:"options"`
							TranslationKey string   `json:"translation_key"`
						} `json:"select"`
					} `json:"selector"`
				} `json:"hvac_mode"`
			} `json:"fields"`
		} `json:"set_hvac_mode"`
		SetPresetMode struct {
			Fields struct {
				PresetMode struct {
					Example string `json:"example"`

					Required bool `json:"required"`
					Selector struct {
						Text interface{} `json:"text"`
					} `json:"selector"`
				} `json:"preset_mode"`
			} `json:"fields"`
		} `json:"set_preset_mode"`
		SetSwingMode struct {
			Fields struct {
				SwingMode struct {
					Example string `json:"example"`

					Required bool `json:"required"`
					Selector struct {
						Text interface{} `json:"text"`
					} `json:"selector"`
				} `json:"swing_mode"`
			} `json:"fields"`
		} `json:"set_swing_mode"`
		SetTemperature struct {
			Fields struct {
				HvacMode struct {
					Selector struct {
						Select struct {
							Options        []string `json:"options"`
							TranslationKey string   `json:"translation_key"`
						} `json:"select"`
					} `json:"selector"`
				} `json:"hvac_mode"`
				TargetTempHigh struct {
					Advanced bool `json:"advanced"`
					Filter   struct {
						SupportedFeatures []int `json:"supported_features"`
					} `json:"filter"`

					Selector struct {
						Number struct {
							Max  int     `json:"max"`
							Min  int     `json:"min"`
							Mode string  `json:"mode"`
							Step float64 `json:"step"`
						} `json:"number"`
					} `json:"selector"`
				} `json:"target_temp_high"`
				TargetTempLow struct {
					Advanced bool `json:"advanced"`
					Filter   struct {
						SupportedFeatures []int `json:"supported_features"`
					} `json:"filter"`
					Selector struct {
						Number struct {
							Max  int     `json:"max"`
							Min  int     `json:"min"`
							Mode string  `json:"mode"`
							Step float64 `json:"step"`
						} `json:"number"`
					} `json:"selector"`
				} `json:"target_temp_low"`
				Temperature struct {
					Filter struct {
						SupportedFeatures []int `json:"supported_features"`
					} `json:"filter"`

					Selector struct {
						Number struct {
							Max  int     `json:"max"`
							Min  int     `json:"min"`
							Mode string  `json:"mode"`
							Step float64 `json:"step"`
						} `json:"number"`
					} `json:"selector"`
				} `json:"temperature"`
			} `json:"fields"`
		} `json:"set_temperature"`
		Toggle  struct{} `json:"toggle"`
		TurnOff struct{} `json:"turn_off"`
		TurnOn  struct{} `json:"turn_on"`
	} `json:"services"`
}

type Fan struct {
	Domain   string `json:"domain"`
	Services struct {
		DecreaseSpeed struct {
			Fields struct {
				PercentageStep struct {
					Advanced bool `json:"advanced"`

					Required bool `json:"required"`
					Selector struct {
						Number struct {
							Max               int    `json:"max"`
							Min               int    `json:"min"`
							UnitOfMeasurement string `json:"unit_of_measurement"`
						} `json:"number"`
					} `json:"selector"`
				} `json:"percentage_step"`
			} `json:"fields"`
		} `json:"decrease_speed"`
		IncreaseSpeed struct {
			Fields struct {
				PercentageStep struct {
					Advanced bool `json:"advanced"`

					Required bool `json:"required"`
					Selector struct {
						Number struct {
							Max               int    `json:"max"`
							Min               int    `json:"min"`
							UnitOfMeasurement string `json:"unit_of_measurement"`
						} `json:"number"`
					} `json:"selector"`
				} `json:"percentage_step"`
			} `json:"fields"`
		} `json:"increase_speed"`
		Oscillate struct {
			Fields struct {
				Oscillating struct {
					Required bool `json:"required"`
					Selector struct {
						Boolean interface{} `json:"boolean"`
					} `json:"selector"`
				} `json:"oscillating"`
			} `json:"fields"`
		} `json:"oscillate"`
		SetDirection struct {
			Fields struct {
				Direction struct {
					Required bool `json:"required"`
					Selector struct {
						Select struct {
							Options        []string `json:"options"`
							TranslationKey string   `json:"translation_key"`
						} `json:"select"`
					} `json:"selector"`
				} `json:"direction"`
			} `json:"fields"`
		} `json:"set_direction"`
		SetPercentage struct {
			Fields struct {
				Percentage struct {
					Required bool `json:"required"`
					Selector struct {
						Number struct {
							Max               int    `json:"max"`
							Min               int    `json:"min"`
							UnitOfMeasurement string `json:"unit_of_measurement"`
						} `json:"number"`
					} `json:"selector"`
				} `json:"percentage"`
			} `json:"fields"`
		} `json:"set_percentage"`
		SetPresetMode struct {
			Fields struct {
				PresetMode struct {
					Example string `json:"example"`

					Required bool `json:"required"`
					Selector struct {
						Text interface{} `json:"text"`
					} `json:"selector"`
				} `json:"preset_mode"`
			} `json:"fields"`
		} `json:"set_preset_mode"`
		Toggle  struct{} `json:"toggle"`
		TurnOff struct{} `json:"turn_off"`
		TurnOn  struct {
			Fields struct {
				Percentage struct {
					Filter struct {
						SupportedFeatures []int `json:"supported_features"`
					} `json:"filter"`

					Selector struct {
						Number struct {
							Max               int    `json:"max"`
							Min               int    `json:"min"`
							UnitOfMeasurement string `json:"unit_of_measurement"`
						} `json:"number"`
					} `json:"selector"`
				} `json:"percentage"`
				PresetMode struct {
					Example string `json:"example"`
					Filter  struct {
						SupportedFeatures []int `json:"supported_features"`
					} `json:"filter"`

					Selector struct {
						Text interface{} `json:"text"`
					} `json:"selector"`
				} `json:"preset_mode"`
			} `json:"fields"`
		} `json:"turn_on"`
	} `json:"services"`
}

type Light struct {
	Domain   string `json:"domain"`
	Services struct {
		Toggle struct {
			Fields struct {
				Brightness struct {
					Advanced bool `json:"advanced"`

					Filter struct {
						Attribute struct {
							SupportedColorModes []string `json:"supported_color_modes"`
						} `json:"attribute"`
					} `json:"filter"`

					Selector struct {
						Number struct {
							Max int `json:"max"`
							Min int `json:"min"`
						} `json:"number"`
					} `json:"selector"`
				} `json:"brightness"`
				BrightnessPct struct {
					Filter struct {
						Attribute struct {
							SupportedColorModes []string `json:"supported_color_modes"`
						} `json:"attribute"`
					} `json:"filter"`

					Selector struct {
						Number struct {
							Max               int    `json:"max"`
							Min               int    `json:"min"`
							UnitOfMeasurement string `json:"unit_of_measurement"`
						} `json:"number"`
					} `json:"selector"`
				} `json:"brightness_pct"`
				ColorName struct {
					Advanced bool `json:"advanced"`
					Filter   struct {
						Attribute struct {
							SupportedColorModes []string `json:"supported_color_modes"`
						} `json:"attribute"`
					} `json:"filter"`

					Selector struct {
						Select struct {
							Options        []string `json:"options"`
							TranslationKey string   `json:"translation_key"`
						} `json:"select"`
					} `json:"selector"`
				} `json:"color_name"`
				ColorTemp struct {
					Advanced bool `json:"advanced"`

					Filter struct {
						Attribute struct {
							SupportedColorModes []string `json:"supported_color_modes"`
						} `json:"attribute"`
					} `json:"filter"`

					Selector struct {
						ColorTemp interface{} `json:"color_temp"`
					} `json:"selector"`
				} `json:"color_temp"`
				Effect struct {
					Filter struct {
						SupportedFeatures []int `json:"supported_features"`
					} `json:"filter"`

					Selector struct {
						Text interface{} `json:"text"`
					} `json:"selector"`
				} `json:"effect"`
				Flash struct {
					Advanced bool `json:"advanced"`

					Filter struct {
						SupportedFeatures []int `json:"supported_features"`
					} `json:"filter"`

					Selector struct {
						Select struct {
							Options []struct {
								Label string `json:"label"`
								Value string `json:"value"`
							} `json:"options"`
						} `json:"select"`
					} `json:"selector"`
				} `json:"flash"`
				HsColor struct {
					Advanced bool `json:"advanced"`

					Example string `json:"example"`
					Filter  struct {
						Attribute struct {
							SupportedColorModes []string `json:"supported_color_modes"`
						} `json:"attribute"`
					} `json:"filter"`

					Selector struct {
						Object interface{} `json:"object"`
					} `json:"selector"`
				} `json:"hs_color"`
				Kelvin struct {
					Advanced bool `json:"advanced"`

					Filter struct {
						Attribute struct {
							SupportedColorModes []string `json:"supported_color_modes"`
						} `json:"attribute"`
					} `json:"filter"`

					Selector struct {
						ColorTemp struct {
							Max  int    `json:"max"`
							Min  int    `json:"min"`
							Unit string `json:"unit"`
						} `json:"color_temp"`
					} `json:"selector"`
				} `json:"kelvin"`
				Profile struct {
					Advanced bool `json:"advanced"`

					Example string `json:"example"`

					Selector struct {
						Text interface{} `json:"text"`
					} `json:"selector"`
				} `json:"profile"`
				RgbColor struct {
					Advanced bool `json:"advanced"`

					Example string `json:"example"`
					Filter  struct {
						Attribute struct {
							SupportedColorModes []string `json:"supported_color_modes"`
						} `json:"attribute"`
					} `json:"filter"`

					Selector struct {
						ColorRgb interface{} `json:"color_rgb"`
					} `json:"selector"`
				} `json:"rgb_color"`
				Transition struct {
					Filter struct {
						SupportedFeatures []int `json:"supported_features"`
					} `json:"filter"`

					Selector struct {
						Number struct {
							Max               int    `json:"max"`
							Min               int    `json:"min"`
							UnitOfMeasurement string `json:"unit_of_measurement"`
						} `json:"number"`
					} `json:"selector"`
				} `json:"transition"`
				White struct {
					Advanced bool `json:"advanced"`

					Filter struct {
						Attribute struct {
							SupportedColorModes []string `json:"supported_color_modes"`
						} `json:"attribute"`
					} `json:"filter"`

					Selector struct {
						Constant struct {
							Label string `json:"label"`
							Value bool   `json:"value"`
						} `json:"constant"`
					} `json:"selector"`
				} `json:"white"`
				XyColor struct {
					Advanced bool `json:"advanced"`

					Example string `json:"example"`
					Filter  struct {
						Attribute struct {
							SupportedColorModes []string `json:"supported_color_modes"`
						} `json:"attribute"`
					} `json:"filter"`

					Selector struct {
						Object interface{} `json:"object"`
					} `json:"selector"`
				} `json:"xy_color"`
			} `json:"fields"`
			Name   string `json:"name"`
			Target struct {
				Entity []struct {
					Domain []string `json:"domain"`
				} `json:"entity"`
			} `json:"target"`
		} `json:"toggle"`
		TurnOff struct {
			Fields struct {
				Flash struct {
					Advanced bool `json:"advanced"`

					Filter struct {
						SupportedFeatures []int `json:"supported_features"`
					} `json:"filter"`

					Selector struct {
						Select struct {
							Options []struct {
								Label string `json:"label"`
								Value string `json:"value"`
							} `json:"options"`
						} `json:"select"`
					} `json:"selector"`
				} `json:"flash"`
				Transition struct {
					Filter struct {
						SupportedFeatures []int `json:"supported_features"`
					} `json:"filter"`

					Selector struct {
						Number struct {
							Max               int    `json:"max"`
							Min               int    `json:"min"`
							UnitOfMeasurement string `json:"unit_of_measurement"`
						} `json:"number"`
					} `json:"selector"`
				} `json:"transition"`
			} `json:"fields"`
			Name   string `json:"name"`
			Target struct {
				Entity []struct {
					Domain []string `json:"domain"`
				} `json:"entity"`
			} `json:"target"`
		} `json:"turn_off"`
		TurnOn struct {
			Fields struct {
				Brightness struct {
					Advanced bool `json:"advanced"`

					Filter struct {
						Attribute struct {
							SupportedColorModes []string `json:"supported_color_modes"`
						} `json:"attribute"`
					} `json:"filter"`

					Selector struct {
						Number struct {
							Max int `json:"max"`
							Min int `json:"min"`
						} `json:"number"`
					} `json:"selector"`
				} `json:"brightness"`
				BrightnessPct struct {
					Filter struct {
						Attribute struct {
							SupportedColorModes []string `json:"supported_color_modes"`
						} `json:"attribute"`
					} `json:"filter"`

					Selector struct {
						Number struct {
							Max               int    `json:"max"`
							Min               int    `json:"min"`
							UnitOfMeasurement string `json:"unit_of_measurement"`
						} `json:"number"`
					} `json:"selector"`
				} `json:"brightness_pct"`
				BrightnessStep struct {
					Advanced bool `json:"advanced"`

					Filter struct {
						Attribute struct {
							SupportedColorModes []string `json:"supported_color_modes"`
						} `json:"attribute"`
					} `json:"filter"`

					Selector struct {
						Number struct {
							Max int `json:"max"`
							Min int `json:"min"`
						} `json:"number"`
					} `json:"selector"`
				} `json:"brightness_step"`
				BrightnessStepPct struct {
					Filter struct {
						Attribute struct {
							SupportedColorModes []string `json:"supported_color_modes"`
						} `json:"attribute"`
					} `json:"filter"`

					Selector struct {
						Number struct {
							Max               int    `json:"max"`
							Min               int    `json:"min"`
							UnitOfMeasurement string `json:"unit_of_measurement"`
						} `json:"number"`
					} `json:"selector"`
				} `json:"brightness_step_pct"`
				ColorName struct {
					Advanced bool `json:"advanced"`

					Filter struct {
						Attribute struct {
							SupportedColorModes []string `json:"supported_color_modes"`
						} `json:"attribute"`
					} `json:"filter"`

					Selector struct {
						Select struct {
							Options        []string `json:"options"`
							TranslationKey string   `json:"translation_key"`
						} `json:"select"`
					} `json:"selector"`
				} `json:"color_name"`
				ColorTemp struct {
					Filter struct {
						Attribute struct {
							SupportedColorModes []string `json:"supported_color_modes"`
						} `json:"attribute"`
					} `json:"filter"`

					Selector struct {
						ColorTemp struct {
							Max  int    `json:"max"`
							Min  int    `json:"min"`
							Unit string `json:"unit"`
						} `json:"color_temp"`
					} `json:"selector"`
				} `json:"color_temp"`
				Effect struct {
					Filter struct {
						SupportedFeatures []int `json:"supported_features"`
					} `json:"filter"`

					Selector struct {
						Text interface{} `json:"text"`
					} `json:"selector"`
				} `json:"effect"`
				Flash struct {
					Advanced bool `json:"advanced"`

					Filter struct {
						SupportedFeatures []int `json:"supported_features"`
					} `json:"filter"`

					Selector struct {
						Select struct {
							Options []struct {
								Label string `json:"label"`
								Value string `json:"value"`
							} `json:"options"`
						} `json:"select"`
					} `json:"selector"`
				} `json:"flash"`
				HsColor struct {
					Advanced bool `json:"advanced"`

					Example string `json:"example"`
					Filter  struct {
						Attribute struct {
							SupportedColorModes []string `json:"supported_color_modes"`
						} `json:"attribute"`
					} `json:"filter"`

					Selector struct {
						Object interface{} `json:"object"`
					} `json:"selector"`
				} `json:"hs_color"`
				Kelvin struct {
					Advanced bool `json:"advanced"`

					Filter struct {
						Attribute struct {
							SupportedColorModes []string `json:"supported_color_modes"`
						} `json:"attribute"`
					} `json:"filter"`

					Selector struct {
						ColorTemp struct {
							Max  int    `json:"max"`
							Min  int    `json:"min"`
							Unit string `json:"unit"`
						} `json:"color_temp"`
					} `json:"selector"`
				} `json:"kelvin"`
				Profile struct {
					Advanced bool `json:"advanced"`

					Example string `json:"example"`

					Selector struct {
						Text interface{} `json:"text"`
					} `json:"selector"`
				} `json:"profile"`
				RgbColor struct {
					Filter struct {
						Attribute struct {
							SupportedColorModes []string `json:"supported_color_modes"`
						} `json:"attribute"`
					} `json:"filter"`

					Selector struct {
						ColorRgb interface{} `json:"color_rgb"`
					} `json:"selector"`
				} `json:"rgb_color"`
				RgbwColor struct {
					Advanced bool `json:"advanced"`

					Example string `json:"example"`
					Filter  struct {
						Attribute struct {
							SupportedColorModes []string `json:"supported_color_modes"`
						} `json:"attribute"`
					} `json:"filter"`

					Selector struct {
						Object interface{} `json:"object"`
					} `json:"selector"`
				} `json:"rgbw_color"`
				RgbwwColor struct {
					Advanced bool `json:"advanced"`

					Example string `json:"example"`
					Filter  struct {
						Attribute struct {
							SupportedColorModes []string `json:"supported_color_modes"`
						} `json:"attribute"`
					} `json:"filter"`

					Selector struct {
						Object interface{} `json:"object"`
					} `json:"selector"`
				} `json:"rgbww_color"`
				Transition struct {
					Filter struct {
						SupportedFeatures []int `json:"supported_features"`
					} `json:"filter"`

					Selector struct {
						Number struct {
							Max               int    `json:"max"`
							Min               int    `json:"min"`
							UnitOfMeasurement string `json:"unit_of_measurement"`
						} `json:"number"`
					} `json:"selector"`
				} `json:"transition"`
				White struct {
					Advanced bool `json:"advanced"`

					Filter struct {
						Attribute struct {
							SupportedColorModes []string `json:"supported_color_modes"`
						} `json:"attribute"`
					} `json:"filter"`

					Selector struct {
						Constant struct {
							Label string `json:"label"`
							Value bool   `json:"value"`
						} `json:"constant"`
					} `json:"selector"`
				} `json:"white"`
				XyColor struct {
					Advanced bool `json:"advanced"`

					Example string `json:"example"`
					Filter  struct {
						Attribute struct {
							SupportedColorModes []string `json:"supported_color_modes"`
						} `json:"attribute"`
					} `json:"filter"`

					Selector struct {
						Object interface{} `json:"object"`
					} `json:"selector"`
				} `json:"xy_color"`
			} `json:"fields"`
			Name   string `json:"name"`
			Target struct {
				Entity []struct {
					Domain []string `json:"domain"`
				} `json:"entity"`
			} `json:"target"`
		} `json:"turn_on"`
	} `json:"services"`
}

type Switch struct {
	Domain   string `json:"domain"`
	Services struct {
		Toggle struct {
			Fields struct {
			} `json:"fields"`
			Name   string `json:"name"`
			Target struct {
				Entity []struct {
					Domain []string `json:"domain"`
				} `json:"entity"`
			} `json:"target"`
		} `json:"toggle"`
		TurnOff struct {
			Fields struct {
			} `json:"fields"`
			Name   string `json:"name"`
			Target struct {
				Entity []struct {
					Domain []string `json:"domain"`
				} `json:"entity"`
			} `json:"target"`
		} `json:"turn_off"`
		TurnOn struct {
			Fields struct {
			} `json:"fields"`
			Name   string `json:"name"`
			Target struct {
				Entity []struct {
					Domain []string `json:"domain"`
				} `json:"entity"`
			} `json:"target"`
		} `json:"turn_on"`
	} `json:"services"`
}

type XiaomiMiot struct {
	Domain   string `json:"domain"`
	Services struct {
		CallAction struct {
			Fields struct {
				Aiid struct {
					Example  int64 `json:"example"`
					Required bool  `json:"required"`
					Selector struct {
						Number struct {
							Max  int64  `json:"max"`
							Min  int64  `json:"min"`
							Mode string `json:"mode"`
						} `json:"number"`
					} `json:"selector"`
				} `json:"aiid"`
				EntityID struct {
					Example string `json:"example"`

					Required bool `json:"required"`
					Selector struct {
						Entity struct {
							Integration string `json:"integration"`
						} `json:"entity"`
					} `json:"selector"`
				} `json:"entity_id"`
				Params struct {
					Example  string `json:"example"`
					Selector struct {
						Object interface{} `json:"object"`
					} `json:"selector"`
				} `json:"params"`
				Siid struct {
					Example  int64 `json:"example"`
					Required bool  `json:"required"`
					Selector struct {
						Number struct {
							Max  int64  `json:"max"`
							Min  int64  `json:"min"`
							Mode string `json:"mode"`
						} `json:"number"`
					} `json:"selector"`
				} `json:"siid"`
			} `json:"fields"`

			Response struct {
				Optional bool `json:"optional"`
			} `json:"response"`
		} `json:"call_action"`
		GetBindkey struct {
			Fields struct {
				Did struct {
					Selector struct {
						Text interface{} `json:"text"`
					} `json:"selector"`
				} `json:"did"`
				EntityID struct {
					Example string `json:"example"`

					Required bool `json:"required"`
					Selector struct {
						Entity struct {
							Integration string `json:"integration"`
						} `json:"entity"`
					} `json:"selector"`
				} `json:"entity_id"`
			} `json:"fields"`

			Response struct {
				Optional bool `json:"optional"`
			} `json:"response"`
		} `json:"get_bindkey"`
		GetDeviceData struct {
			Fields struct {
				EntityID struct {
					Example string `json:"example"`

					Required bool `json:"required"`
					Selector struct {
						Entity struct {
							Integration string `json:"integration"`
						} `json:"entity"`
					} `json:"selector"`
				} `json:"entity_id"`
				Group struct {
					Default string `json:"default"`

					Example string `json:"example"`

					Selector struct {
						Select struct {
							Options []string `json:"options"`
						} `json:"select"`
					} `json:"selector"`
				} `json:"group"`
				Key struct {
					Example string `json:"example"`

					Required bool `json:"required"`
					Selector struct {
						Text interface{} `json:"text"`
					} `json:"selector"`
				} `json:"key"`
				Limit struct {
					Example int64 `json:"example"`

					Selector struct {
						Number struct {
							Max  int64  `json:"max"`
							Min  int64  `json:"min"`
							Mode string `json:"mode"`
						} `json:"number"`
					} `json:"selector"`
				} `json:"limit"`
				TimeEnd struct {
					Example int64 `json:"example"`

					Selector struct {
						Number struct {
							Max  int64  `json:"max"`
							Min  int64  `json:"min"`
							Mode string `json:"mode"`
						} `json:"number"`
					} `json:"selector"`
				} `json:"time_end"`
				TimeStart struct {
					Example int64 `json:"example"`

					Selector struct {
						Number struct {
							Max  int64  `json:"max"`
							Min  int64  `json:"min"`
							Mode string `json:"mode"`
						} `json:"number"`
					} `json:"selector"`
				} `json:"time_start"`
				Type struct {
					Default string `json:"default"`

					Example string `json:"example"`

					Selector struct {
						Select struct {
							Options []string `json:"options"`
						} `json:"select"`
					} `json:"selector"`
				} `json:"type"`
			} `json:"fields"`

			Response struct {
				Optional bool `json:"optional"`
			} `json:"response"`
		} `json:"get_device_data"`
		GetProperties struct {
			Fields struct {
				EntityID struct {
					Example string `json:"example"`

					Required bool `json:"required"`
					Selector struct {
						Entity struct {
							Integration string `json:"integration"`
						} `json:"entity"`
					} `json:"selector"`
				} `json:"entity_id"`
				Mapping struct {
					Example string `json:"example"`

					Required bool `json:"required"`
					Selector struct {
						Object interface{} `json:"object"`
					} `json:"selector"`
				} `json:"mapping"`
				UpdateEntity struct {
					Default bool `json:"default"`

					Example bool `json:"example"`

					Selector struct {
						Boolean interface{} `json:"boolean"`
					} `json:"selector"`
				} `json:"update_entity"`
			} `json:"fields"`

			Response struct {
				Optional bool `json:"optional"`
			} `json:"response"`
		} `json:"get_properties"`
		GetToken struct {
			Fields struct {
				Name struct {
					Required bool `json:"required"`
					Selector struct {
						Text interface{} `json:"text"`
					} `json:"selector"`
				} `json:"name"`
			} `json:"fields"`

			Response struct {
				Optional bool `json:"optional"`
			} `json:"response"`
		} `json:"get_token"`
		IntelligentSpeaker struct {
			Fields struct {
				EntityID struct {
					Example string `json:"example"`

					Required bool `json:"required"`
					Selector struct {
						Entity struct {
							Domain      string `json:"domain"`
							Integration string `json:"integration"`
						} `json:"entity"`
					} `json:"selector"`
				} `json:"entity_id"`
				Execute struct {
					Default bool `json:"default"`

					Example bool `json:"example"`

					Selector struct {
						Boolean interface{} `json:"boolean"`
					} `json:"selector"`
				} `json:"execute"`
				Silent struct {
					Default bool `json:"default"`

					Example bool `json:"example"`

					Selector struct {
						Boolean interface{} `json:"boolean"`
					} `json:"selector"`
				} `json:"silent"`
				Text struct {
					Example string `json:"example"`

					Required bool `json:"required"`
					Selector struct {
						Text interface{} `json:"text"`
					} `json:"selector"`
				} `json:"text"`
			} `json:"fields"`

			Response struct {
				Optional bool `json:"optional"`
			} `json:"response"`
		} `json:"intelligent_speaker"`
		Reload struct {
			Description string   `json:"description"`
			Fields      struct{} `json:"fields"`
			Name        string   `json:"name"`
		} `json:"reload"`
		RenewDevices struct {
			Fields struct {
				Username struct {
					Required bool `json:"required"`
					Selector struct {
						Text interface{} `json:"text"`
					} `json:"selector"`
				} `json:"username"`
			} `json:"fields"`
			Name string `json:"name"`
		} `json:"renew_devices"`
		RequestXiaomiAPI struct {
			Fields struct {
				API struct {
					Example string `json:"example"`

					Required bool `json:"required"`
					Selector struct {
						Text interface{} `json:"text"`
					} `json:"selector"`
				} `json:"api"`
				Crypt struct {
					Default bool `json:"default"`

					Example bool `json:"example"`

					Selector struct {
						Boolean interface{} `json:"boolean"`
					} `json:"selector"`
				} `json:"crypt"`
				Data struct {
					Example string `json:"example"`

					Selector struct {
						Object interface{} `json:"object"`
					} `json:"selector"`
				} `json:"data"`
				EntityID struct {
					Example string `json:"example"`

					Required bool `json:"required"`
					Selector struct {
						Entity struct {
							Integration string `json:"integration"`
						} `json:"entity"`
					} `json:"selector"`
				} `json:"entity_id"`
				Method struct {
					Default string `json:"default"`

					Example string `json:"example"`

					Selector struct {
						Select struct {
							Options []string `json:"options"`
						} `json:"select"`
					} `json:"selector"`
				} `json:"method"`
				Sid struct {
					Default string `json:"default"`

					Example  string `json:"example"`
					Selector struct {
						Select struct {
							Options []struct {
								Label string `json:"label"`
								Value string `json:"value"`
							} `json:"options"`
						} `json:"select"`
					} `json:"selector"`
				} `json:"sid"`
			} `json:"fields"`

			Response struct {
				Optional bool `json:"optional"`
			} `json:"response"`
		} `json:"request_xiaomi_api"`
		SendCommand struct {
			Fields struct {
				EntityID struct {
					Example string `json:"example"`

					Required bool `json:"required"`
					Selector struct {
						Entity struct {
							Integration string `json:"integration"`
						} `json:"entity"`
					} `json:"selector"`
				} `json:"entity_id"`
				Method struct {
					Example string `json:"example"`

					Required bool `json:"required"`
					Selector struct {
						Text interface{} `json:"text"`
					} `json:"selector"`
				} `json:"method"`
				Params struct {
					Description string   `json:"description"`
					Example     []string `json:"example"`
					Name        string   `json:"name"`
					Required    bool     `json:"required"`
					Selector    struct {
						Object interface{} `json:"object"`
					} `json:"selector"`
				} `json:"params"`
			} `json:"fields"`

			Response struct {
				Optional bool `json:"optional"`
			} `json:"response"`
		} `json:"send_command"`
		SetMiotProperty struct {
			Fields struct {
				EntityID struct {
					Example string `json:"example"`

					Required bool `json:"required"`
					Selector struct {
						Entity struct {
							Integration string `json:"integration"`
						} `json:"entity"`
					} `json:"selector"`
				} `json:"entity_id"`
				Piid struct {
					Example  int64 `json:"example"`
					Required bool  `json:"required"`
					Selector struct {
						Number struct {
							Max  int64  `json:"max"`
							Min  int64  `json:"min"`
							Mode string `json:"mode"`
						} `json:"number"`
					} `json:"selector"`
				} `json:"piid"`
				Siid struct {
					Example  int64 `json:"example"`
					Required bool  `json:"required"`
					Selector struct {
						Number struct {
							Max  int64  `json:"max"`
							Min  int64  `json:"min"`
							Mode string `json:"mode"`
						} `json:"number"`
					} `json:"selector"`
				} `json:"siid"`
				Value struct {
					Example  bool `json:"example"`
					Required bool `json:"required"`
					Selector struct {
						Object interface{} `json:"object"`
					} `json:"selector"`
				} `json:"value"`
			} `json:"fields"`

			Response struct {
				Optional bool `json:"optional"`
			} `json:"response"`
		} `json:"set_miot_property"`
		SetProperty struct {
			Fields struct {
				EntityID struct {
					Example string `json:"example"`

					Required bool `json:"required"`
					Selector struct {
						Entity struct {
							Integration string `json:"integration"`
						} `json:"entity"`
					} `json:"selector"`
				} `json:"entity_id"`
				Field struct {
					Example string `json:"example"`

					Required bool `json:"required"`
					Selector struct {
						Text interface{} `json:"text"`
					} `json:"selector"`
				} `json:"field"`
				Value struct {
					Example bool `json:"example"`

					Required bool `json:"required"`
					Selector struct {
						Object interface{} `json:"object"`
					} `json:"selector"`
				} `json:"value"`
			} `json:"fields"`

			Response struct {
				Optional bool `json:"optional"`
			} `json:"response"`
		} `json:"set_property"`
		XiaoaiWakeup struct {
			Fields struct {
				EntityID struct {
					Example string `json:"example"`

					Required bool `json:"required"`
					Selector struct {
						Entity struct {
							Domain      string `json:"domain"`
							Integration string `json:"integration"`
						} `json:"entity"`
					} `json:"selector"`
				} `json:"entity_id"`
				Text struct {
					Example string `json:"example"`

					Selector struct {
						Text interface{} `json:"text"`
					} `json:"selector"`
				} `json:"text"`
			} `json:"fields"`

			Response struct {
				Optional bool `json:"optional"`
			} `json:"response"`
		} `json:"xiaoai_wakeup"`
	} `json:"services"`
}
