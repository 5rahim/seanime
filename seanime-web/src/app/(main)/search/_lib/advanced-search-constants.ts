import { AL_MediaFormat } from "@/api/generated/types"

export const ADVANCED_SEARCH_MEDIA_GENRES = [
    "Action",
    "Adventure",
    "Comedy",
    "Drama",
    "Ecchi",
    "Fantasy",
    "Horror",
    "Mahou Shoujo",
    "Mecha",
    "Music",
    "Mystery",
    "Psychological",
    "Romance",
    "Sci-Fi",
    "Slice of Life",
    "Sports",
    "Supernatural",
    "Thriller",
]

export const ADVANCED_SEARCH_SEASONS = [
    "Winter",
    "Spring",
    "Summer",
    "Fall",
]

export const ADVANCED_SEARCH_FORMATS: { value: AL_MediaFormat, label: string }[] = [
    { value: "TV", label: "TV" },
    { value: "MOVIE", label: "Movie" },
    { value: "ONA", label: "ONA" },
    { value: "OVA", label: "OVA" },
    { value: "TV_SHORT", label: "TV Short" },
    { value: "SPECIAL", label: "Special" },
]

export const ADVANCED_SEARCH_FORMATS_MANGA: { value: AL_MediaFormat, label: string }[] = [
    { value: "MANGA", label: "Manga" },
    { value: "ONE_SHOT", label: "One Shot" },
]


export const ADVANCED_SEARCH_COUNTRIES_MANGA: { value: string, label: string }[] = [
    { value: "JP", label: "Japan" },
    { value: "KR", label: "South Korea" },
    { value: "CN", label: "China" },
    { value: "TW", label: "Taiwan" },
]

export const ADVANCED_SEARCH_STATUS = [
    { value: "FINISHED", label: "Finished" },
    { value: "RELEASING", label: "Releasing" },
    { value: "NOT_YET_RELEASED", label: "Upcoming" },
    { value: "HIATUS", label: "Hiatus" },
    { value: "CANCELLED", label: "Cancelled" },
]

export const ADVANCED_SEARCH_SORTING = [
    { value: "TRENDING_DESC", label: "Trending" },
    { value: "START_DATE_DESC", label: "Release date" },
    { value: "SCORE_DESC", label: "Highest score" },
    { value: "POPULARITY_DESC", label: "Most popular" },
    { value: "EPISODES_DESC", label: "Number of episodes" },
]

export const ADVANCED_SEARCH_SORTING_MANGA = [
    { value: "TRENDING_DESC", label: "Trending" },
    { value: "START_DATE_DESC", label: "Release date" },
    { value: "SCORE_DESC", label: "Highest score" },
    { value: "POPULARITY_DESC", label: "Most popular" },
    { value: "CHAPTERS_DESC", label: "Number of chapters" },
]

export const ADVANCED_SEARCH_TYPE = [
    { value: "anime", label: "Anime" },
    { value: "manga", label: "Manga" },
]

// Query used to get all tags: 
//  query Query {
//   MediaTagCollection {
//     id,
//     name,
//     description,
//     category,
//     isAdult
//   }
// }

export const ADVANCED_SEARCH_MEDIA_TAGS = [
    {
        "id": 206,
        "name": "4-koma",
        "description": "A manga in the 'yonkoma' format, which consists of four equal-sized panels arranged in a vertical strip.",
        "category": "Technical",
        "isAdult": false
    },
    {
        "id": 710,
        "name": "Achromatic",
        "description": "Contains animation that is primarily done in black and white.",
        "category": "Technical",
        "isAdult": false
    },
    {
        "id": 156,
        "name": "Achronological Order",
        "description": "Chapters or episodes do not occur in chronological order.",
        "category": "Setting-Time",
        "isAdult": false
    },
    {
        "id": 1612,
        "name": "Acrobatics",
        "description": "The art of jumping, tumbling, and balancing. Often paired with trapeze, trampolining, tightropes, or general gymnastics.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 548,
        "name": "Acting",
        "description": "Centers around actors or the acting industry.",
        "category": "Theme-Arts",
        "isAdult": false
    },
    {
        "id": 1052,
        "name": "Adoption",
        "description": "Features a character who has been adopted by someone who is neither of their biological parents.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 505,
        "name": "Advertisement",
        "description": "Produced in order to promote the products of a certain company.",
        "category": "Technical",
        "isAdult": false
    },
    {
        "id": 147,
        "name": "Afterlife",
        "description": "Partly or completely set in the afterlife.",
        "category": "Setting-Universe",
        "isAdult": false
    },
    {
        "id": 138,
        "name": "Age Gap",
        "description": "Prominently features romantic relations between people with a significant age difference.",
        "category": "Theme-Romance",
        "isAdult": false
    },
    {
        "id": 488,
        "name": "Age Regression",
        "description": "Prominently features a character who was returned to a younger state.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 334,
        "name": "Agender",
        "description": "Prominently features agender characters.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 909,
        "name": "Agriculture",
        "description": "Prominently features agriculture practices.",
        "category": "Theme-Slice of Life",
        "isAdult": false
    },
    {
        "id": 279,
        "name": "Ahegao",
        "description": "Features a character making an exaggerated orgasm face.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 167,
        "name": "Airsoft",
        "description": "Centers around the sport of airsoft.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 1291,
        "name": "Alchemy",
        "description": "Features character(s) who practice alchemy.",
        "category": "Theme-Fantasy",
        "isAdult": false
    },
    {
        "id": 191,
        "name": "Aliens",
        "description": "Prominently features extraterrestrial lifeforms.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 146,
        "name": "Alternate Universe",
        "description": "Features multiple alternate universes in the same series.",
        "category": "Setting-Universe",
        "isAdult": false
    },
    {
        "id": 314,
        "name": "American Football",
        "description": "Centers around the sport of American football.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 240,
        "name": "Amnesia",
        "description": "Prominently features a character(s) with memory loss.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 652,
        "name": "Amputation",
        "description": "Features amputation or amputees.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 490,
        "name": "Anachronism",
        "description": "Prominently features elements that are out of place in the historical period the work takes place in, particularly modern elements in a historical setting.",
        "category": "Setting-Time",
        "isAdult": false
    },
    {
        "id": 185,
        "name": "Anal Sex",
        "description": "Features sexual penetration of the anal cavity.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 1598,
        "name": "Ancient China",
        "description": "Setting in ancient china, does not apply to fantasy settings.",
        "category": "Setting-Time",
        "isAdult": false
    },
    {
        "id": 1068,
        "name": "Angels",
        "description": "Prominently features spiritual beings usually represented with wings and halos and believed to be attendants of God.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 433,
        "name": "Animals",
        "description": "Prominently features animal characters in a leading role.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 471,
        "name": "Anthology",
        "description": "A collection of separate works collated into a single release.",
        "category": "Technical",
        "isAdult": false
    },
    {
        "id": 1304,
        "name": "Anthropomorphism",
        "description": "Contains non-human character(s) that have attributes or characteristics of a human being.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 104,
        "name": "Anti-Hero",
        "description": "Features a protagonist who lacks conventional heroic attributes and may be considered a borderline villain.",
        "category": "Cast-Main Cast",
        "isAdult": false
    },
    {
        "id": 251,
        "name": "Archery",
        "description": "Centers around the sport of archery, or prominently features the use of archery in combat.",
        "category": "Theme-Action",
        "isAdult": false
    },
    {
        "id": 629,
        "name": "Armpits",
        "description": "Features the sexual depiction or stimulation of a character's armpits.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 1681,
        "name": "Aromantic",
        "description": "Features a character who experiences little to no romantic attraction.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 1578,
        "name": "Arranged Marriage",
        "description": "Features two characters made to marry each other, usually by their family.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 517,
        "name": "Artificial Intelligence",
        "description": "Intelligent non-organic machines that work and react similarly to humans.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 622,
        "name": "Asexual",
        "description": "Features a character who isn't sexually attracted to people of any sex or gender.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 315,
        "name": "Ashikoki",
        "description": "Footjob; features stimulation of genitalia by feet.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 651,
        "name": "Asphyxiation",
        "description": "Features breath play.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 322,
        "name": "Assassins",
        "description": "Centers around characters who murder people as a profession.",
        "category": "Theme-Other-Organisations",
        "isAdult": false
    },
    {
        "id": 623,
        "name": "Astronomy",
        "description": "Relating or centered around the study of celestial objects and phenomena, space, or the universe.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 264,
        "name": "Athletics",
        "description": "Centers around sporting events that involve competitive running, jumping, throwing, or walking.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 364,
        "name": "Augmented Reality",
        "description": "Prominently features events with augmented reality as the main setting.",
        "category": "Setting-Universe",
        "isAdult": false
    },
    {
        "id": 519,
        "name": "Autobiographical",
        "description": "Real stories and anecdotes written by the author about their own life.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 355,
        "name": "Aviation",
        "description": "Regarding the flying or operation of aircraft.",
        "category": "Theme-Other-Vehicle",
        "isAdult": false
    },
    {
        "id": 235,
        "name": "Badminton",
        "description": "Centers around the sport of badminton.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 1867,
        "name": "Ballet",
        "description": "Prominently features the dance art of ballet. Both traditional and contemporary styles.",
        "category": "Theme-Arts",
        "isAdult": false
    },
    {
        "id": 110,
        "name": "Band",
        "description": "Main cast is a group of musicians.",
        "category": "Theme-Arts-Music",
        "isAdult": false
    },
    {
        "id": 161,
        "name": "Bar",
        "description": "Partly or completely set in a bar.",
        "category": "Setting-Scene",
        "isAdult": false
    },
    {
        "id": 149,
        "name": "Baseball",
        "description": "Centers around the sport of baseball.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 192,
        "name": "Basketball",
        "description": "Centers around the sport of basketball.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 101,
        "name": "Battle Royale",
        "description": "Centers around a fierce group competition, often violent and with only one winner.",
        "category": "Theme-Action",
        "isAdult": false
    },
    {
        "id": 239,
        "name": "Biographical",
        "description": "Based on true stories of real persons living or dead, written by another.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 294,
        "name": "Bisexual",
        "description": "Features a character who is romantically or sexually attracted to people of more than one sex or gender.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 309,
        "name": "Blackmail",
        "description": "Features a character blackmailing another.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 1492,
        "name": "Board Game",
        "description": "Centers around characters playing board games.",
        "category": "Theme-Game",
        "isAdult": false
    },
    {
        "id": 1489,
        "name": "Boarding School",
        "description": "Features characters attending a boarding school.",
        "category": "Setting-Scene",
        "isAdult": false
    },
    {
        "id": 639,
        "name": "Body Horror",
        "description": "Features characters who undergo horrific transformations or disfigurement, often to their own detriment.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 1779,
        "name": "Body Image",
        "description": "Features themes of self-esteem concerning perceived defects or flaws in appearance, such as body weight or disfigurement, and may discuss topics such as eating disorders, fatphobia, and body dysmorphia.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 154,
        "name": "Body Swapping",
        "description": "Centers around individuals swapping bodies with one another.",
        "category": "Theme-Fantasy",
        "isAdult": false
    },
    {
        "id": 246,
        "name": "Bondage",
        "description": "Features BDSM, with or without the use of accessories.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 126,
        "name": "Boobjob",
        "description": "Features the stimulation of male genitalia by breasts.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 1632,
        "name": "Bowling",
        "description": "Centers around the sport of Bowling.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 169,
        "name": "Boxing",
        "description": "Centers around the sport of boxing.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 75,
        "name": "Boys' Love",
        "description": "Prominently features romance between two males, not inherently sexual.",
        "category": "Theme-Romance",
        "isAdult": false
    },
    {
        "id": 171,
        "name": "Bullying",
        "description": "Prominently features the use of force for intimidation, often in a school setting.",
        "category": "Theme-Drama",
        "isAdult": false
    },
    {
        "id": 812,
        "name": "Butler",
        "description": "Prominently features a character who is a butler.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 204,
        "name": "Calligraphy",
        "description": "Centers around the art of calligraphy.",
        "category": "Theme-Arts",
        "isAdult": false
    },
    {
        "id": 1724,
        "name": "Camping",
        "description": "Features the recreational activity of camping, either in a tent, vehicle, or simply sleeping outdoors.",
        "category": "Setting-Scene",
        "isAdult": false
    },
    {
        "id": 870,
        "name": "Cannibalism",
        "description": "Prominently features the act of consuming another member of the same species as food.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 178,
        "name": "Card Battle",
        "description": "Centers around individuals competing in card games.",
        "category": "Theme-Game-Card & Board Game",
        "isAdult": false
    },
    {
        "id": 10,
        "name": "Cars",
        "description": "Centers around the use of automotive vehicles.",
        "category": "Theme-Other-Vehicle",
        "isAdult": false
    },
    {
        "id": 632,
        "name": "Centaur",
        "description": "Prominently features a character with a human upper body and the lower body of a horse.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 1573,
        "name": "Cervix Penetration",
        "description": "A sexual act in which the cervix is visibly penetrated.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 90,
        "name": "CGI",
        "description": "Prominently features scenes created with computer-generated imagery.",
        "category": "Technical",
        "isAdult": false
    },
    {
        "id": 1709,
        "name": "Cheating",
        "description": "Features a character with a partner shown being intimate with someone else consensually.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 646,
        "name": "Cheerleading",
        "description": "Centers around the activity of cheerleading.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 324,
        "name": "Chibi",
        "description": "Features \"super deformed\" character designs with smaller, rounder proportions and a cute look.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 774,
        "name": "Chimera",
        "description": "Features a beast made by combining animals, usually with humans.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 267,
        "name": "Chuunibyou",
        "description": "Prominently features a character with \"Middle School 2nd Year Syndrome\", who either acts like a know-it-all adult or falsely believes they have special powers.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 476,
        "name": "Circus",
        "description": "Prominently features a circus.",
        "category": "Setting-Scene",
        "isAdult": false
    },
    {
        "id": 1403,
        "name": "Class Struggle",
        "description": "Contains conflict born between the different social classes. Generally between an dominant elite and a suffering oppressed group.",
        "category": "Theme-Drama",
        "isAdult": false
    },
    {
        "id": 227,
        "name": "Classic Literature",
        "description": "Discusses or adapts a work of classic world literature.",
        "category": "Theme-Arts",
        "isAdult": false
    },
    {
        "id": 1673,
        "name": "Classical Music",
        "description": "Centers on the musical style of classical, not to be applied to anime that use classical in its soundtrack.",
        "category": "Theme-Arts-Music",
        "isAdult": false
    },
    {
        "id": 1342,
        "name": "Clone",
        "description": "Prominently features a character who is an artificial exact copy of another organism.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 1537,
        "name": "Coastal",
        "description": "Story prominently takes place near the beach or around a coastal area/town. Setting is near the ocean.",
        "category": "Setting-Scene",
        "isAdult": false
    },
    {
        "id": 1753,
        "name": "Cohabitation",
        "description": "Features two or more people who live in the same household and develop a romantic or sexual relationship.",
        "category": "Theme-Romance",
        "isAdult": false
    },
    {
        "id": 404,
        "name": "College",
        "description": "Partly or completely set in a college or university.",
        "category": "Setting-Scene",
        "isAdult": false
    },
    {
        "id": 102,
        "name": "Coming of Age",
        "description": "Centers around a character's transition from childhood to adulthood.",
        "category": "Theme-Drama",
        "isAdult": false
    },
    {
        "id": 456,
        "name": "Conspiracy",
        "description": "Contains one or more factions controlling or attempting to control the world from the shadows.",
        "category": "Theme-Drama",
        "isAdult": false
    },
    {
        "id": 636,
        "name": "Cosmic Horror",
        "description": "A type of horror that emphasizes human insignificance in the grand scope of cosmic reality; fearing the unknown and being powerless to fight it.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 328,
        "name": "Cosplay",
        "description": "Features dressing up as a different character or profession.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 1648,
        "name": "Cowboys",
        "description": "Features Western or Western-inspired cowboys.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 1771,
        "name": "Creature Taming",
        "description": "Features the taming of animals, monsters, or other creatures.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 648,
        "name": "Crime",
        "description": "Centers around unlawful activities punishable by the state or other authority.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 1589,
        "name": "Criminal Organization",
        "description": "Prominently features a group of people who commit crimes for illicit or violent purposes.",
        "category": "Theme-Other-Organisations",
        "isAdult": false
    },
    {
        "id": 186,
        "name": "Crossdressing",
        "description": "Prominently features a character dressing up as the opposite sex.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 158,
        "name": "Crossover",
        "description": "Centers around the placement of two or more otherwise discrete fictional characters, settings, or universes into the context of a single story.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 586,
        "name": "Cult",
        "description": "Features a social group with unorthodox religious, spiritual, or philosophical beliefs and practices.",
        "category": "Theme-Other-Organisations",
        "isAdult": false
    },
    {
        "id": 326,
        "name": "Cultivation",
        "description": "Features characters using training, often martial arts-related, and other special methods to cultivate qi (a component of traditional Chinese philosophy, described as \"life force\") and gain strength or immortality.",
        "category": "Theme-Fantasy",
        "isAdult": false
    },
    {
        "id": 1273,
        "name": "Cumflation",
        "description": "The stomach area expands outward like a balloon due to being filled specifically with semen.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 135,
        "name": "Cunnilingus",
        "description": "Features oral sex performed on female genitalia.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 1689,
        "name": "Curses",
        "description": "Features a character, object or area that has been cursed, usually by a malevolent supernatural force.",
        "category": "Theme-Fantasy",
        "isAdult": false
    },
    {
        "id": 1063,
        "name": "Cute Boys Doing Cute Things",
        "description": "Centers around male characters doing cute activities, usually with little to no emphasis on drama and conflict.",
        "category": "Theme-Slice of Life",
        "isAdult": false
    },
    {
        "id": 92,
        "name": "Cute Girls Doing Cute Things",
        "description": "Centers around female characters doing cute activities, usually with little to no emphasis on drama and conflict.\n",
        "category": "Theme-Slice of Life",
        "isAdult": false
    },
    {
        "id": 108,
        "name": "Cyberpunk",
        "description": "Set in a future of advanced technological and scientific achievements that have resulted in social disorder.",
        "category": "Theme-Sci-Fi",
        "isAdult": false
    },
    {
        "id": 801,
        "name": "Cyborg",
        "description": "Prominently features a human character whose physiological functions are aided or enhanced by artificial means.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 151,
        "name": "Cycling",
        "description": "Centers around the sport of cycling.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 209,
        "name": "Dancing",
        "description": "Centers around the art of dance.",
        "category": "Theme-Arts-Music",
        "isAdult": false
    },
    {
        "id": 615,
        "name": "Death Game",
        "description": "Features characters participating in a game, where failure results in death.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 1234,
        "name": "Deepthroat",
        "description": "Features oral sex where the majority of the erect male genitalia is inside another person's mouth, usually stimulating some gagging in the back of their throat.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 129,
        "name": "Defloration",
        "description": "Features a female character who has never had sexual relations (until now).",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 395,
        "name": "Delinquents",
        "description": "Features characters with a notorious image and attitude, sometimes referred to as \"yankees\".",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 15,
        "name": "Demons",
        "description": "Prominently features malevolent otherworldly creatures.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 654,
        "name": "Denpa",
        "description": "Works that feature themes of social dissociation, delusions, and other issues like suicide, bullying, self-isolation, paranoia, and technological necessity in daily lives. Classic iconography: telephone poles, rooftops, and trains.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 1453,
        "name": "Desert",
        "description": "Prominently features a desert environment.",
        "category": "Setting-Scene",
        "isAdult": false
    },
    {
        "id": 694,
        "name": "Detective",
        "description": "Features a character who investigates and solves crimes.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 1038,
        "name": "DILF",
        "description": "Features sexual intercourse with older men.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 704,
        "name": "Dinosaurs",
        "description": "Prominently features Dinosaurs, prehistoric reptiles that went extinct millions of years ago.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 1219,
        "name": "Disability",
        "description": "A work that features one or more characters with a physical, mental, cognitive, or developmental condition that impairs, interferes with, or limits the person's ability to engage in certain tasks or actions.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 536,
        "name": "Dissociative Identities",
        "description": "A case where one or more people share the same body.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 1511,
        "name": "Double Penetration",
        "description": "A sexual act in which the vagina/anus are penetrated by two penises/toys.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 224,
        "name": "Dragons",
        "description": "Prominently features mythical reptiles which generally have wings and can breathe fire.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 118,
        "name": "Drawing",
        "description": "Centers around the art of drawing, including manga and doujinshi.",
        "category": "Theme-Arts",
        "isAdult": false
    },
    {
        "id": 489,
        "name": "Drugs",
        "description": "Prominently features the usage of drugs such as opioids, stimulants, hallucinogens etc.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 658,
        "name": "Dullahan",
        "description": "Prominently features a character who is a Dullahan, a creature from Irish Folklore with a head that can be detached from its main body.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 604,
        "name": "Dungeon",
        "description": "Prominently features a dungeon environment.",
        "category": "Setting-Scene",
        "isAdult": false
    },
    {
        "id": 217,
        "name": "Dystopian",
        "description": "Partly or completely set in a society characterized by poverty, squalor or oppression.",
        "category": "Setting-Time",
        "isAdult": false
    },
    {
        "id": 792,
        "name": "E-Sports",
        "description": "Prominently features professional video game competitions, tournaments, players, etc.",
        "category": "Theme-Game",
        "isAdult": false
    },
    {
        "id": 1667,
        "name": "Eco-Horror",
        "description": "Utilizes a horrifying depiction of ecology to explore man and its relationship with nature.",
        "category": "Theme-Drama",
        "isAdult": false
    },
    {
        "id": 248,
        "name": "Economics",
        "description": "Centers around the field of economics.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 140,
        "name": "Educational",
        "description": "Primary aim is to educate the audience.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 1666,
        "name": "Elderly Protagonist",
        "description": "The protagonist is either over 60 years of age, has an elderly appearance, or, in the case of non-humans, is considered elderly for their species.",
        "category": "Cast-Main Cast",
        "isAdult": false
    },
    {
        "id": 598,
        "name": "Elf",
        "description": "Prominently features a character who is an elf.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 105,
        "name": "Ensemble Cast",
        "description": "Features a large cast of characters with (almost) equal screen time and importance to the plot.",
        "category": "Cast-Main Cast",
        "isAdult": false
    },
    {
        "id": 342,
        "name": "Environmental",
        "description": "Concern with the state of the natural world and how humans interact with it.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 193,
        "name": "Episodic",
        "description": "Features story arcs that are loosely tied or lack an overarching plot.",
        "category": "Technical",
        "isAdult": false
    },
    {
        "id": 482,
        "name": "Ero Guro",
        "description": "Japanese literary and artistic movement originating in the 1930's. Works have a focus on grotesque eroticism, sexual corruption, and decadence.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 1722,
        "name": "Erotic Piercings",
        "description": "Features a type of body modification designed to enhance sexual pleasure and intimacy, and/or decoratively adorns portions of the body considered sexual in nature.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 310,
        "name": "Espionage",
        "description": "Prominently features characters infiltrating an organization in order to steal sensitive information.",
        "category": "Theme-Action",
        "isAdult": false
    },
    {
        "id": 1564,
        "name": "Estranged Family",
        "description": "At least one family member of the MC intentionally distances themselves or a family distances themselves from a person related to the MC.",
        "category": "Cast-Main Cast",
        "isAdult": false
    },
    {
        "id": 901,
        "name": "Exhibitionism",
        "description": "Features the act of exposing oneself in public for sexual pleasure.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 1719,
        "name": "Exorcism",
        "description": "Involving religious methods of vanquishing youkai, demons, or other supernatural entities.",
        "category": "Theme-Fantasy",
        "isAdult": false
    },
    {
        "id": 130,
        "name": "Facial",
        "description": "Features sexual ejaculation onto an individual's face.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 1455,
        "name": "Fairy",
        "description": "Prominently features a character who is a fairy.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 400,
        "name": "Fairy Tale",
        "description": "This work tells a fairy tale, centers around fairy tales, or is based on a classic fairy tale.",
        "category": "Theme-Fantasy",
        "isAdult": false
    },
    {
        "id": 1657,
        "name": "Fake Relationship",
        "description": "When two characters enter a fake relationship that mutually benefits one or both involved.",
        "category": "Theme-Drama",
        "isAdult": false
    },
    {
        "id": 87,
        "name": "Family Life",
        "description": "Centers around the activities of a family unit.",
        "category": "Theme-Slice of Life",
        "isAdult": false
    },
    {
        "id": 410,
        "name": "Fashion",
        "description": "Centers around the fashion industry.",
        "category": "Theme-Arts",
        "isAdult": false
    },
    {
        "id": 631,
        "name": "Feet",
        "description": "Features the sexual depiction or stimulation of a character's feet.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 134,
        "name": "Fellatio",
        "description": "Blowjob; features oral sex performed on male genitalia.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 23,
        "name": "Female Harem",
        "description": "Main cast features the protagonist plus several female characters who are romantically interested in them.",
        "category": "Theme-Romance",
        "isAdult": false
    },
    {
        "id": 98,
        "name": "Female Protagonist",
        "description": "Main character is female.",
        "category": "Cast-Main Cast",
        "isAdult": false
    },
    {
        "id": 1456,
        "name": "Femboy",
        "description": "Features a boy who exhibits characteristics or behaviors considered in many cultures to be typical of girls.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 1146,
        "name": "Femdom",
        "description": "Female Dominance. Features sexual acts with a woman in a dominant position.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 1132,
        "name": "Fencing",
        "description": "Centers around the sport of fencing.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 1422,
        "name": "Filmmaking",
        "description": "Centers around the art of filmmaking.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 1807,
        "name": "Fingering",
        "description": "Features vaginal or anal insertion of fingers.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 613,
        "name": "Firefighters",
        "description": "Centered around the life and activities of rescuers specialised in firefighting.",
        "category": "Theme-Other-Organisations",
        "isAdult": false
    },
    {
        "id": 212,
        "name": "Fishing",
        "description": "Centers around the sport of fishing.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 1659,
        "name": "Fisting",
        "description": "A sexual activity that involves inserting one or more hands into the vagina or rectum.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 170,
        "name": "Fitness",
        "description": "Centers around exercise with the aim of improving physical health.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 242,
        "name": "Flash",
        "description": "Created using Flash animation techniques.",
        "category": "Technical",
        "isAdult": false
    },
    {
        "id": 136,
        "name": "Flat Chest",
        "description": "Features a female character with smaller-than-average breasts.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 142,
        "name": "Food",
        "description": "Centers around cooking or food appraisal.",
        "category": "Theme-Arts",
        "isAdult": false
    },
    {
        "id": 148,
        "name": "Football",
        "description": "Centers around the sport of football (known in the USA as \"soccer\").",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 198,
        "name": "Foreign",
        "description": "Partly or completely set in a country outside the country of origin.",
        "category": "Setting-Scene",
        "isAdult": false
    },
    {
        "id": 1277,
        "name": "Found Family",
        "description": "Features a group of characters with no biological relations that are united in a group providing social support.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 226,
        "name": "Fugitive",
        "description": "Prominently features a character evading capture by an individual or organization.",
        "category": "Theme-Action",
        "isAdult": false
    },
    {
        "id": 89,
        "name": "Full CGI",
        "description": "Almost entirely created with computer-generated imagery.",
        "category": "Technical",
        "isAdult": false
    },
    {
        "id": 207,
        "name": "Full Color",
        "description": "Manga that were initially published in full color.",
        "category": "Technical",
        "isAdult": false
    },
    {
        "id": 188,
        "name": "Futanari",
        "description": "Features female characters with male genitalia.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 91,
        "name": "Gambling",
        "description": "Centers around the act of gambling.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 106,
        "name": "Gangs",
        "description": "Centers around gang organizations.",
        "category": "Theme-Other-Organisations",
        "isAdult": false
    },
    {
        "id": 77,
        "name": "Gender Bending",
        "description": "Prominently features a character who dresses and behaves in a way characteristic of another gender, or has been transformed into a person of another gender.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 220,
        "name": "Ghost",
        "description": "Prominently features a character who is a ghost.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 443,
        "name": "Go",
        "description": "Centered around the game of Go.",
        "category": "Theme-Game-Card & Board Game",
        "isAdult": false
    },
    {
        "id": 480,
        "name": "Goblin",
        "description": "A goblin is a monstrous creature from European folklore. They are almost always small and grotesque, mischievous or outright malicious, and greedy. Sometimes with magical abilities.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 253,
        "name": "Gods",
        "description": "Prominently features a character of divine or religious nature.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 532,
        "name": "Golf",
        "description": "Centers around the sport of golf.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 94,
        "name": "Gore",
        "description": "Prominently features graphic bloodshed and violence.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 830,
        "name": "Group Sex",
        "description": "Features more than two participants engaged in sex simultaneously.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 157,
        "name": "Guns",
        "description": "Prominently features the use of guns in combat.",
        "category": "Theme-Action",
        "isAdult": false
    },
    {
        "id": 356,
        "name": "Gyaru",
        "description": "Prominently features a female character who has a distinct American-emulated fashion style, such as tanned skin, bleached hair, and excessive makeup. Also known as gal.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 1591,
        "name": "Hair Pulling",
        "description": "A sexual act in which the giver will grab the receivers hair and tug whilst giving pleasure from behind.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 1396,
        "name": "Handball",
        "description": "Centers around the sport of handball.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 317,
        "name": "Handjob",
        "description": "Features the stimulation of genitalia by another's hands.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 99,
        "name": "Henshin",
        "description": "Prominently features character or costume transformations which often grant special abilities.",
        "category": "Theme-Fantasy",
        "isAdult": false
    },
    {
        "id": 1045,
        "name": "Heterosexual",
        "description": "Prominently features a romance between a man and a woman, not inherently sexual.",
        "category": "Theme-Romance",
        "isAdult": false
    },
    {
        "id": 282,
        "name": "Hikikomori",
        "description": "Prominently features a character who withdraws from social life, often seeking extreme isolation.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 1615,
        "name": "Hip-hop Music",
        "description": "Centers on the musical style of hip-hop, not to be applied to anime that use hip-hop in its soundtrack.",
        "category": "Theme-Arts-Music",
        "isAdult": false
    },
    {
        "id": 25,
        "name": "Historical",
        "description": "Partly or completely set during a real period of world history.",
        "category": "Setting-Time",
        "isAdult": false
    },
    {
        "id": 1379,
        "name": "Homeless",
        "description": "Prominently features a character that is homeless.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 1525,
        "name": "Horticulture",
        "description": "The story prominently features plant care and gardening.",
        "category": "Theme-Slice of Life",
        "isAdult": false
    },
    {
        "id": 181,
        "name": "Human Pet",
        "description": "Features characters in a master-slave relationship where one is the \"owner\" and the other is a \"pet.\"",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 1367,
        "name": "Hypersexuality",
        "description": "Portrays a character with a hypersexuality disorder, compulsive sexual behavior, or sex addiction.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 228,
        "name": "Ice Skating",
        "description": "Centers around the sport of ice skating.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 115,
        "name": "Idol",
        "description": "Centers around the life and activities of an idol.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 128,
        "name": "Incest",
        "description": "Features sexual or romantic relations between characters who are related by blood.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 1788,
        "name": "Indigenous Cultures",
        "description": "Prominently features real-life indigenous cultures.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 1523,
        "name": "Inn",
        "description": "Partially or completely set in an Inn or Hotel.",
        "category": "Setting-Scene",
        "isAdult": false
    },
    {
        "id": 1102,
        "name": "Inseki",
        "description": "Features sexual or romantic relations among step, adopted, and other non-blood related family members.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 359,
        "name": "Irrumatio",
        "description": "Oral rape; features a character thrusting their genitalia or a phallic object into the mouth of another character.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 244,
        "name": "Isekai",
        "description": "Features characters being transported into an alternate world setting and having to adapt to their new surroundings.",
        "category": "Theme-Fantasy",
        "isAdult": false
    },
    {
        "id": 81,
        "name": "Iyashikei",
        "description": "Primary aim is to heal the audience through serene depictions of characters' daily lives.",
        "category": "Theme-Slice of Life",
        "isAdult": false
    },
    {
        "id": 1672,
        "name": "Jazz Music",
        "description": "Centers on the musical style of jazz, not to be applied to anime that use jazz in its soundtrack.",
        "category": "Theme-Arts-Music",
        "isAdult": false
    },
    {
        "id": 27,
        "name": "Josei",
        "description": "Target demographic is adult females.",
        "category": "Demographic",
        "isAdult": false
    },
    {
        "id": 1105,
        "name": "Judo",
        "description": "Centers around the sport of judo.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 1866,
        "name": "Kabuki",
        "description": "Prominently features the traditional Japanese theater art of kabuki.",
        "category": "Theme-Arts",
        "isAdult": false
    },
    {
        "id": 276,
        "name": "Kaiju",
        "description": "Prominently features giant monsters.",
        "category": "Theme-Fantasy",
        "isAdult": false
    },
    {
        "id": 182,
        "name": "Karuta",
        "description": "Centers around the game of karuta.",
        "category": "Theme-Game-Card & Board Game",
        "isAdult": false
    },
    {
        "id": 254,
        "name": "Kemonomimi",
        "description": "Prominently features humanoid characters with animal ears.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 28,
        "name": "Kids",
        "description": "Target demographic is young children.",
        "category": "Demographic",
        "isAdult": false
    },
    {
        "id": 1419,
        "name": "Kingdom Management",
        "description": "Characters in these series take on the responsibility of running a town or kingdom, whether they take control of an existing one, or build their own from the ground up.",
        "category": "Theme-Drama",
        "isAdult": false
    },
    {
        "id": 1607,
        "name": "Konbini",
        "description": "Predominantly features a convenience store.",
        "category": "Setting-Scene",
        "isAdult": false
    },
    {
        "id": 779,
        "name": "Kuudere",
        "description": "Prominently features a character who generally retains a cold, blunt and cynical exterior, but once one gets to know them, they have a very warm and loving interior.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 437,
        "name": "Lacrosse",
        "description": "A team game played with a ball and lacrosse sticks.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 650,
        "name": "Lactation",
        "description": "Features breast milk play and production.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 516,
        "name": "Language Barrier",
        "description": "A barrier to communication between people who are unable to speak a common language.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 137,
        "name": "Large Breasts",
        "description": "Features a character with larger-than-average breasts.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 483,
        "name": "LGBTQ+ Themes",
        "description": "Prominently features characters or themes associated with the LGBTQ+ community, such as sexuality or gender identity.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 1789,
        "name": "Long Strip",
        "description": "Manga originally published in a vertical, long-strip format, designed for viewing on smartphones. Also known as webtoons.",
        "category": "Technical",
        "isAdult": false
    },
    {
        "id": 466,
        "name": "Lost Civilization",
        "description": "Featuring a civilization with few ruins or records that exist in present day knowledge.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 139,
        "name": "Love Triangle",
        "description": "Centered around romantic feelings between more than two people. Includes all love polygons.",
        "category": "Theme-Romance",
        "isAdult": false
    },
    {
        "id": 107,
        "name": "Mafia",
        "description": "Centered around Italian organised crime syndicates.",
        "category": "Theme-Other-Organisations",
        "isAdult": false
    },
    {
        "id": 29,
        "name": "Magic",
        "description": "Prominently features magical elements or the use of magic.",
        "category": "Theme-Fantasy",
        "isAdult": false
    },
    {
        "id": 117,
        "name": "Mahjong",
        "description": "Centered around the game of mahjong.",
        "category": "Theme-Game-Card & Board Game",
        "isAdult": false
    },
    {
        "id": 249,
        "name": "Maids",
        "description": "Prominently features a character who is a maid.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 1140,
        "name": "Makeup",
        "description": "Centers around the makeup industry.",
        "category": "Theme-Arts",
        "isAdult": false
    },
    {
        "id": 123,
        "name": "Male Harem",
        "description": "Main cast features the protagonist plus several male characters who are romantically interested in them.",
        "category": "Theme-Romance",
        "isAdult": false
    },
    {
        "id": 1450,
        "name": "Male Pregnancy",
        "description": "Features pregnant male characters in a sexual context.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 82,
        "name": "Male Protagonist",
        "description": "Main character is male.",
        "category": "Cast-Main Cast",
        "isAdult": false
    },
    {
        "id": 1835,
        "name": "Manzai",
        "description": "Prominently features an act of traditional Japanese comedy that involves two performers.",
        "category": "Theme-Arts",
        "isAdult": false
    },
    {
        "id": 1433,
        "name": "Marriage",
        "description": "Centers around marriage between two or more characters.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 30,
        "name": "Martial Arts",
        "description": "Centers around the use of traditional hand-to-hand combat.",
        "category": "Theme-Action",
        "isAdult": false
    },
    {
        "id": 999,
        "name": "Masochism",
        "description": "Prominently features characters who get sexual pleasure from being hurt or controlled by others.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 131,
        "name": "Masturbation",
        "description": "Features erotic stimulation of one's own genitalia or other erogenous regions.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 1624,
        "name": "Matchmaking",
        "description": "Prominently features either a matchmaker or events with the intent of matchmaking with eventual marriage in sight.",
        "category": "Theme-Romance",
        "isAdult": false
    },
    {
        "id": 1653,
        "name": "Mating Press",
        "description": "Features the sex position in which two partners face each other, with one of them thrusting downwards and the other's legs tucked up towards their head.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 1526,
        "name": "Matriarchy",
        "description": "Prominently features a country that is ruled by a Queen or a society that is dominated by female inheritance.",
        "category": "Setting",
        "isAdult": false
    },
    {
        "id": 659,
        "name": "Medicine",
        "description": "Centered around the activities of people working in the field of medicine.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 1752,
        "name": "Medieval",
        "description": "Partially or completely set in the Middle Ages or a Middle Ages-inspired setting. Commonly features elements such as European castles and knights.",
        "category": "Setting-Time",
        "isAdult": false
    },
    {
        "id": 365,
        "name": "Memory Manipulation",
        "description": "Prominently features a character(s) who has had their memories altered.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 558,
        "name": "Mermaid",
        "description": "A mythological creature with the body of a human and the tail of a fish.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 144,
        "name": "Meta",
        "description": "Features fourth wall-breaking references to itself or genre tropes.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 1675,
        "name": "Metal Music",
        "description": "Centers on the musical style of metal, not to be applied to anime that use metal in its soundtrack.",
        "category": "Theme-Arts-Music",
        "isAdult": false
    },
    {
        "id": 187,
        "name": "MILF",
        "description": "Features sexual intercourse with older women.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 34,
        "name": "Military",
        "description": "Centered around the life and activities of military personnel.",
        "category": "Theme-Other-Organisations",
        "isAdult": false
    },
    {
        "id": 1361,
        "name": "Mixed Gender Harem",
        "description": "Main cast features the protagonist plus several people, regardless of gender, who are romantically interested in them.",
        "category": "Theme-Romance",
        "isAdult": false
    },
    {
        "id": 1742,
        "name": "Mixed Media",
        "description": "Features a combination of different media and animation techniques. Often seen with puppetry, textiles, live action footage, stop motion, and more. This does not include works with normal usage of CGI in their production.",
        "category": "Technical",
        "isAdult": false
    },
    {
        "id": 1848,
        "name": "Modeling",
        "description": "Features a line of work with the purpose of displaying and advertising products such as makeup, clothing, and jewelry. Also includes posing artistically for figure drawing, painting, sculpting, and photography.",
        "category": "Theme-Arts",
        "isAdult": false
    },
    {
        "id": 1090,
        "name": "Monster Boy",
        "description": "Prominently features a male character who is a part-monster.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 259,
        "name": "Monster Girl",
        "description": "Prominently features a female character who is part-monster.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 176,
        "name": "Mopeds",
        "description": "Prominently features mopeds.",
        "category": "Theme-Other-Vehicle",
        "isAdult": false
    },
    {
        "id": 173,
        "name": "Motorcycles",
        "description": "Prominently features the use of motorcycles.",
        "category": "Theme-Other-Vehicle",
        "isAdult": false
    },
    {
        "id": 1544,
        "name": "Mountaineering",
        "description": "Prominently features characters discussing or hiking mountains.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 265,
        "name": "Musical Theater",
        "description": "Features a performance that combines songs, spoken dialogue, acting, and dance.",
        "category": "Theme-Arts-Music",
        "isAdult": false
    },
    {
        "id": 208,
        "name": "Mythology",
        "description": "Prominently features mythological elements, especially those from religious or cultural tradition.",
        "category": "Theme-Fantasy",
        "isAdult": false
    },
    {
        "id": 125,
        "name": "Nakadashi",
        "description": "Creampie; features sexual ejaculation inside of a character.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 1501,
        "name": "Natural Disaster",
        "description": "It focuses on catastrophic events of natural origin, such as earthquakes, tsunamis,  volcanic eruptions, and severe storms. These works often present situations of extreme danger in which the characters struggle to survive and overcome the adversity.",
        "category": "Setting-Scene",
        "isAdult": false
    },
    {
        "id": 1243,
        "name": "Necromancy",
        "description": "When the dead are summoned as spirits, skeletons, or the undead, usually for the purpose of gaining information or to be used as a weapon.",
        "category": "Theme-Fantasy",
        "isAdult": false
    },
    {
        "id": 113,
        "name": "Nekomimi",
        "description": "Humanoid characters with cat-like features such as cat ears and a tail.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 124,
        "name": "Netorare",
        "description": "Netorare is what happens when the protagonist gets their partner stolen from them by someone else. It is a sexual fetish designed to cause sexual jealousy by way of having the partner indulge in sexual activity with someone other than the protagonist.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 280,
        "name": "Netorase",
        "description": "Features characters in a romantic relationship who agree to be sexually intimate with others.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 316,
        "name": "Netori",
        "description": "Features the protagonist stealing the partner of someone else. The opposite of netorare.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 255,
        "name": "Ninja",
        "description": "Prominently features Japanese warriors traditionally trained in espionage, sabotage and assasination.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 341,
        "name": "No Dialogue",
        "description": "This work contains no dialogue.",
        "category": "Technical",
        "isAdult": false
    },
    {
        "id": 327,
        "name": "Noir",
        "description": "Stylized as a cynical crime drama with low-key visuals.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 1337,
        "name": "Non-fiction",
        "description": "A work that provides information regarding a real world topic and does not focus on an imaginary narrative.",
        "category": "Technical",
        "isAdult": false
    },
    {
        "id": 100,
        "name": "Nudity",
        "description": "Features a character wearing no clothing or exposing intimate body parts.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 614,
        "name": "Nun",
        "description": "Prominently features a character who is a nun.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 1571,
        "name": "Office",
        "description": "Features people who work in a business office.",
        "category": "Setting-Scene",
        "isAdult": false
    },
    {
        "id": 653,
        "name": "Office Lady",
        "description": "Prominently features a female office worker or OL.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 593,
        "name": "Oiran",
        "description": "Prominently features a courtesan character of the Japanese Edo Period.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 780,
        "name": "Ojou-sama",
        "description": "Features a wealthy, high-class, oftentimes stuck up and demanding female character.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 533,
        "name": "Omegaverse",
        "description": "Alternative universe that prominently features dynamics modeled after wolves in which there are alphas, betas, and omegas and heat cycles as well as impregnation, regardless of gender.",
        "category": "Setting-Universe",
        "isAdult": true
    },
    {
        "id": 1382,
        "name": "Orphan",
        "description": "Prominently features a character that is an orphan.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 97,
        "name": "Otaku Culture",
        "description": "Centers around the culture of a fanatical fan-base.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 262,
        "name": "Outdoor Activities",
        "description": "Centers around hiking, camping or other outdoor activities.",
        "category": "Setting-Scene",
        "isAdult": false
    },
    {
        "id": 1903,
        "name": "Oyakodon",
        "description": "Features a character who has sexual relations with both the mother and her daughter.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 874,
        "name": "Pandemic",
        "description": "Prominently features a disease prevalent over a whole country or the world.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 1645,
        "name": "Parenthood",
        "description": "Centers around the experience of raising a child.",
        "category": "Theme-Slice of Life",
        "isAdult": false
    },
    {
        "id": 949,
        "name": "Parkour",
        "description": "Centers around the sport of parkour.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 39,
        "name": "Parody",
        "description": "Features deliberate exaggeration of popular tropes or a particular genre to comedic effect.",
        "category": "Theme-Comedy",
        "isAdult": false
    },
    {
        "id": 1515,
        "name": "Pet Play",
        "description": "Treating a participant as though they were a pet animal. Often involves a collar and possibly BDSM.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 391,
        "name": "Philosophy",
        "description": "Relating or devoted to the study of the fundamental nature of knowledge, reality, and existence.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 195,
        "name": "Photography",
        "description": "Centers around the use of cameras to capture photos.",
        "category": "Theme-Arts",
        "isAdult": false
    },
    {
        "id": 201,
        "name": "Pirates",
        "description": "Prominently features sea-faring adventurers branded as criminals by the law.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 183,
        "name": "Poker",
        "description": "Centers around the game of poker or its variations.",
        "category": "Theme-Game-Card & Board Game",
        "isAdult": false
    },
    {
        "id": 40,
        "name": "Police",
        "description": "Centers around the life and activities of law enforcement officers.",
        "category": "Theme-Other-Organisations",
        "isAdult": false
    },
    {
        "id": 103,
        "name": "Politics",
        "description": "Centers around politics, politicians, or government activities.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 1459,
        "name": "Polyamorous",
        "description": "Features a character who is in a consenting relationship with multiple people at one time.",
        "category": "Theme-Romance",
        "isAdult": false
    },
    {
        "id": 93,
        "name": "Post-Apocalyptic",
        "description": "Partly or completely set in a world or civilization after a global disaster.",
        "category": "Setting-Universe",
        "isAdult": false
    },
    {
        "id": 215,
        "name": "POV",
        "description": "Point of View; features scenes shown from the perspective of the series protagonist.",
        "category": "Technical",
        "isAdult": false
    },
    {
        "id": 397,
        "name": "Pregnancy",
        "description": "Features pregnant female characters or discusses the topic of pregnancy.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 109,
        "name": "Primarily Adult Cast",
        "description": "Main cast is mostly composed of characters above a high school age.",
        "category": "Cast-Main Cast",
        "isAdult": false
    },
    {
        "id": 1601,
        "name": "Primarily Animal Cast",
        "description": "Main cast is mostly composed animal or animal-like characters.",
        "category": "Cast-Main Cast",
        "isAdult": false
    },
    {
        "id": 446,
        "name": "Primarily Child Cast",
        "description": "Main cast is mostly composed of characters below a high school age.",
        "category": "Cast-Main Cast",
        "isAdult": false
    },
    {
        "id": 86,
        "name": "Primarily Female Cast",
        "description": "Main cast is mostly composed of female characters.",
        "category": "Cast-Main Cast",
        "isAdult": false
    },
    {
        "id": 88,
        "name": "Primarily Male Cast",
        "description": "Main cast is mostly composed of male characters.",
        "category": "Cast-Main Cast",
        "isAdult": false
    },
    {
        "id": 1228,
        "name": "Primarily Teen Cast",
        "description": "Main cast is mostly composed of teen characters.",
        "category": "Cast-Main Cast",
        "isAdult": false
    },
    {
        "id": 1427,
        "name": "Prison",
        "description": "Partly or completely set in a prison.",
        "category": "Setting-Scene",
        "isAdult": false
    },
    {
        "id": 190,
        "name": "Prostitution",
        "description": "Features characters who are paid for sexual favors.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 1490,
        "name": "Proxy Battle",
        "description": "A proxy battle is a battle where humans use creatures/robots to do the fighting for them, either by commanding those creatures/robots or by simply evolving them/changing them into battle mode.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 1670,
        "name": "Psychosexual",
        "description": "Work that involves the psychological aspects of sexual impulses.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 127,
        "name": "Public Sex",
        "description": "Features sexual acts performed in public settings.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 325,
        "name": "Puppetry",
        "description": "Animation style involving the manipulation of puppets to act out scenes.",
        "category": "Technical",
        "isAdult": false
    },
    {
        "id": 481,
        "name": "Rakugo",
        "description": "Rakugo is the traditional Japanese performance art of comic storytelling.",
        "category": "Theme-Arts",
        "isAdult": false
    },
    {
        "id": 155,
        "name": "Rape",
        "description": "Features non-consensual sexual penetration.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 160,
        "name": "Real Robot",
        "description": "Prominently features mechanical designs loosely influenced by real-world robotics.",
        "category": "Theme-Sci-Fi-Mecha",
        "isAdult": false
    },
    {
        "id": 311,
        "name": "Rehabilitation",
        "description": "Prominently features the recovery of a character who became incapable of social life or work.",
        "category": "Theme-Drama",
        "isAdult": false
    },
    {
        "id": 243,
        "name": "Reincarnation",
        "description": "Features a character being born again after death, typically as another person or in another world.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 1091,
        "name": "Religion",
        "description": "Centers on the belief that humanity is related to supernatural, transcendental, and spiritual elements.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 1745,
        "name": "Rescue",
        "description": "Centers around operations that carry out urgent treatment of injuries, remove people from danger, or save lives. This includes series that are about search-and-rescue teams, trauma surgeons, firefighters, and more.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 1647,
        "name": "Restaurant",
        "description": "Features a business that prepares and serves food and drinks to customers. Also encompasses cafes and bistros.",
        "category": "Setting-Scene",
        "isAdult": false
    },
    {
        "id": 252,
        "name": "Revenge",
        "description": "Prominently features a character who aims to exact punishment in a resentful or vindictive manner.",
        "category": "Theme-Drama",
        "isAdult": false
    },
    {
        "id": 1907,
        "name": "Reverse Isekai",
        "description": "Features a character from a fantasy world who is transported into a modern day setting.",
        "category": "Theme-Fantasy",
        "isAdult": false
    },
    {
        "id": 845,
        "name": "Rimjob",
        "description": "Features oral sex performed on the anus.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 175,
        "name": "Robots",
        "description": "Prominently features humanoid machines.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 1674,
        "name": "Rock Music",
        "description": "Centers on the musical style of rock, not to be applied to anime that use rock in its soundtrack.",
        "category": "Theme-Arts-Music",
        "isAdult": false
    },
    {
        "id": 683,
        "name": "Rotoscoping",
        "description": "Animation technique that animators use to trace over motion picture footage, frame by frame, to produce realistic action.",
        "category": "Technical",
        "isAdult": false
    },
    {
        "id": 1327,
        "name": "Royal Affairs",
        "description": "Features nobility, alliances, arranged marriage, succession disputes, religious orders and other elements of royal politics.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 221,
        "name": "Rugby",
        "description": "Centers around the sport of rugby.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 250,
        "name": "Rural",
        "description": "Partly or completely set in the countryside.",
        "category": "Setting-Scene",
        "isAdult": false
    },
    {
        "id": 723,
        "name": "Sadism",
        "description": "Prominently features characters deriving pleasure, especially sexual gratification, from inflicting pain, suffering, or humiliation on others.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 291,
        "name": "Samurai",
        "description": "Prominently features warriors of medieval Japanese nobility bound by a code of honor.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 80,
        "name": "Satire",
        "description": "Prominently features the use of comedy or ridicule to expose and criticise social phenomena.",
        "category": "Theme-Comedy",
        "isAdult": false
    },
    {
        "id": 387,
        "name": "Scat",
        "description": "Lots of feces.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 46,
        "name": "School",
        "description": "Partly or completely set in a primary or secondary educational institution.",
        "category": "Setting-Scene",
        "isAdult": false
    },
    {
        "id": 84,
        "name": "School Club",
        "description": "Partly or completely set in a school club scene.",
        "category": "Setting-Scene",
        "isAdult": false
    },
    {
        "id": 1039,
        "name": "Scissoring",
        "description": "A form of sexual activity between women in which the genitals are stimulated by being rubbed against one another.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 811,
        "name": "Scuba Diving",
        "description": "Prominently features characters diving with the aid of special breathing equipment.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 50,
        "name": "Seinen",
        "description": "Target demographic is adult males.",
        "category": "Demographic",
        "isAdult": false
    },
    {
        "id": 133,
        "name": "Sex Toys",
        "description": "Features objects that are designed to stimulate sexual pleasure.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 772,
        "name": "Shapeshifting",
        "description": "Features character(s) who changes one's appearance or form.",
        "category": "Theme-Fantasy",
        "isAdult": false
    },
    {
        "id": 1732,
        "name": "Shimaidon",
        "description": "Features a character who has sexual relations with two sisters.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 305,
        "name": "Ships",
        "description": "Prominently features the use of sea-based transportation vessels.",
        "category": "Theme-Other-Vehicle",
        "isAdult": false
    },
    {
        "id": 121,
        "name": "Shogi",
        "description": "Centers around the game of shogi.",
        "category": "Theme-Game-Card & Board Game",
        "isAdult": false
    },
    {
        "id": 53,
        "name": "Shoujo",
        "description": "Target demographic is teenage and young adult females.",
        "category": "Demographic",
        "isAdult": false
    },
    {
        "id": 56,
        "name": "Shounen",
        "description": "Target demographic is teenage and young adult males.",
        "category": "Demographic",
        "isAdult": false
    },
    {
        "id": 468,
        "name": "Shrine Maiden",
        "description": "Prominently features a character who is a shrine maiden.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 809,
        "name": "Skateboarding",
        "description": "Centers around or prominently features skateboarding as a sport.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 499,
        "name": "Skeleton",
        "description": "Prominently features skeleton(s) as a character.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 83,
        "name": "Slapstick",
        "description": "Prominently features comedy based on deliberately clumsy actions or embarrassing events.",
        "category": "Theme-Comedy",
        "isAdult": false
    },
    {
        "id": 247,
        "name": "Slavery",
        "description": "Prominently features slaves, slavery, or slave trade.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 1616,
        "name": "Snowscape",
        "description": "Prominently or partially set in a snowy environment.",
        "category": "Setting-Scene",
        "isAdult": false
    },
    {
        "id": 386,
        "name": "Software Development",
        "description": "Centers around characters developing or programming a piece of technology, software, gaming, etc.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 63,
        "name": "Space",
        "description": "Partly or completely set in outer space.",
        "category": "Setting-Universe",
        "isAdult": false
    },
    {
        "id": 162,
        "name": "Space Opera",
        "description": "Centers around space warfare, advanced technology, chivalric romance and adventure.",
        "category": "Theme-Sci-Fi",
        "isAdult": false
    },
    {
        "id": 1292,
        "name": "Spearplay",
        "description": "Prominently features the use of spears in combat.",
        "category": "Theme-Action",
        "isAdult": false
    },
    {
        "id": 1235,
        "name": "Squirting",
        "description": "Female ejaculation; features the expulsion of liquid from the female genitalia.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 95,
        "name": "Steampunk",
        "description": "Prominently features technology and designs inspired by 19th-century industrial steam-powered machinery.",
        "category": "Theme-Fantasy",
        "isAdult": false
    },
    {
        "id": 323,
        "name": "Stop Motion",
        "description": "Animation style characterized by physical objects being moved incrementally between frames to create the illusion of movement.",
        "category": "Technical",
        "isAdult": false
    },
    {
        "id": 665,
        "name": "Succubus",
        "description": "Prominently features a character who is a succubus, a creature in medieval folklore that typically uses their sexual prowess to trap and seduce people to feed off them.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 1046,
        "name": "Suicide",
        "description": "The act or an instance of taking or attempting to take one's own life voluntarily and intentionally.",
        "category": "Theme-Drama",
        "isAdult": false
    },
    {
        "id": 360,
        "name": "Sumata",
        "description": "Pussyjob; features the stimulation of male genitalia by the thighs and labia majora of a female character.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 1080,
        "name": "Sumo",
        "description": "Centers around the sport of sumo.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 66,
        "name": "Super Power",
        "description": "Prominently features characters with special abilities that allow them to do what would normally be physically or logically impossible.",
        "category": "Theme-Fantasy",
        "isAdult": false
    },
    {
        "id": 159,
        "name": "Super Robot",
        "description": "Prominently features large robots often piloted by hot-blooded protagonists.",
        "category": "Theme-Sci-Fi-Mecha",
        "isAdult": false
    },
    {
        "id": 172,
        "name": "Superhero",
        "description": "Prominently features super-powered humans who aim to serve the greater good.",
        "category": "Theme-Fantasy",
        "isAdult": false
    },
    {
        "id": 678,
        "name": "Surfing",
        "description": "Centers around surfing as a sport.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 281,
        "name": "Surreal Comedy",
        "description": "Prominently features comedic moments that defy casual reasoning, resulting in illogical events.",
        "category": "Theme-Comedy",
        "isAdult": false
    },
    {
        "id": 143,
        "name": "Survival",
        "description": "Centers around the struggle to live in spite of extreme obstacles.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 1902,
        "name": "Swapping ",
        "description": "Features consensual partner swapping between couples during sexual activities.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 630,
        "name": "Sweat",
        "description": "Lots of sweat.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 150,
        "name": "Swimming",
        "description": "Centers around the sport of swimming.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 43,
        "name": "Swordplay",
        "description": "Prominently features the use of swords in combat.",
        "category": "Theme-Action",
        "isAdult": false
    },
    {
        "id": 120,
        "name": "Table Tennis",
        "description": "Centers around the sport of table tennis (also known as \"ping pong\").",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 174,
        "name": "Tanks",
        "description": "Prominently features the use of tanks or other armoured vehicles.",
        "category": "Theme-Other-Vehicle",
        "isAdult": false
    },
    {
        "id": 335,
        "name": "Tanned Skin",
        "description": "Prominently features characters with tanned skin.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 165,
        "name": "Teacher",
        "description": "Protagonist is an educator, usually in a school setting.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 649,
        "name": "Teens' Love",
        "description": "Sexually explicit love-story between individuals of the opposite sex, specifically targeting females of teens and young adult age.",
        "category": "Theme-Romance",
        "isAdult": false
    },
    {
        "id": 194,
        "name": "Tennis",
        "description": "Centers around the sport of tennis.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 189,
        "name": "Tentacles",
        "description": "Features the long appendages most commonly associated with octopuses or squid, often sexually penetrating a character.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 285,
        "name": "Terrorism",
        "description": "Centers around the activities of a terrorist or terrorist organization.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 132,
        "name": "Threesome",
        "description": "Features sexual acts between three people.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 1596,
        "name": "Time Loop",
        "description": "A character is stuck in a repetitive cycle that they are attempting to break out of. This is distinct from a manipulating time of their own choice.",
        "category": "Theme-Sci-Fi",
        "isAdult": false
    },
    {
        "id": 96,
        "name": "Time Manipulation",
        "description": "Prominently features time-traveling or other time-warping phenomena.",
        "category": "Theme-Sci-Fi",
        "isAdult": false
    },
    {
        "id": 153,
        "name": "Time Skip",
        "description": "Features a gap in time used to advance the story.",
        "category": "Setting-Time",
        "isAdult": false
    },
    {
        "id": 513,
        "name": "Tokusatsu",
        "description": "Prominently features elements that resemble special effects in Japanese live-action shows",
        "category": "Theme-Sci-Fi",
        "isAdult": false
    },
    {
        "id": 931,
        "name": "Tomboy",
        "description": "Features a girl who exhibits characteristics or behaviors considered in many cultures to be typical of boys.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 1121,
        "name": "Torture",
        "description": "The act of deliberately inflicting severe pain or suffering upon another individual or oneself as a punishment or with a specific purpose.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 85,
        "name": "Tragedy",
        "description": "Centers around tragic events and unhappy endings.",
        "category": "Theme-Drama",
        "isAdult": false
    },
    {
        "id": 122,
        "name": "Trains",
        "description": "Prominently features trains.",
        "category": "Theme-Other-Vehicle",
        "isAdult": false
    },
    {
        "id": 1165,
        "name": "Transgender",
        "description": "Features a character whose gender identity differs from the sex they were assigned at birth.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 1310,
        "name": "Travel",
        "description": "Centers around character(s) moving between places a significant distance apart.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 214,
        "name": "Triads",
        "description": "Centered around Chinese organised crime syndicates.",
        "category": "Theme-Other-Organisations",
        "isAdult": false
    },
    {
        "id": 164,
        "name": "Tsundere",
        "description": "Prominently features a character who acts cold and hostile in order to mask warmer emotions.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 494,
        "name": "Twins",
        "description": "Prominently features two or more siblings that were born at one birth.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 1539,
        "name": "Unrequited Love",
        "description": "One or more characters are experiencing an unrequited love that may or may not be reciprocated.",
        "category": "Theme-Romance",
        "isAdult": false
    },
    {
        "id": 595,
        "name": "Urban",
        "description": "Partly or completely set in a city.",
        "category": "Setting-Scene",
        "isAdult": false
    },
    {
        "id": 321,
        "name": "Urban Fantasy",
        "description": "Set in a world similar to the real world, but with the existence of magic or other supernatural elements.",
        "category": "Setting-Universe",
        "isAdult": false
    },
    {
        "id": 74,
        "name": "Vampire",
        "description": "Prominently features a character who is a vampire.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 1790,
        "name": "Vertical Video",
        "description": "Animated works originally created in a vertical aspect ratio (such as 9:16), intended for viewing on smartphones.",
        "category": "Technical",
        "isAdult": false
    },
    {
        "id": 1574,
        "name": "Veterinarian",
        "description": "Prominently features a veterinarian or one of the main characters is a veterinarian.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 308,
        "name": "Video Games",
        "description": "Centers around characters playing video games.",
        "category": "Theme-Game",
        "isAdult": false
    },
    {
        "id": 618,
        "name": "Vikings",
        "description": "Prominently features Scandinavian seafaring pirates and warriors.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 857,
        "name": "Villainess",
        "description": "Centers around or prominently features a villainous noble lady.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 380,
        "name": "Virginity",
        "description": "Features a male character who has never had sexual relations (until now).",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 112,
        "name": "Virtual World",
        "description": "Partly or completely set in the world inside a video game.",
        "category": "Setting-Universe",
        "isAdult": false
    },
    {
        "id": 1690,
        "name": "Vocal Synth",
        "description": "Features one or more singers or characters that are products of a synthesize singing program. Popular examples are Vocaloids, UTAUloids, and CeVIOs.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 116,
        "name": "Volleyball",
        "description": "Centers around the sport of volleyball.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 624,
        "name": "Vore",
        "description": "Features a character being swallowed or swallowing another creature whole.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 318,
        "name": "Voyeur",
        "description": "Features a character who enjoys seeing the sex acts or sex organs of others.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 1047,
        "name": "VTuber",
        "description": "Prominently features a character who is either an actual or fictive VTuber.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 111,
        "name": "War",
        "description": "Partly or completely set during wartime.",
        "category": "Theme-Other",
        "isAdult": false
    },
    {
        "id": 180,
        "name": "Watersports",
        "description": "Features sexual situations involving urine.",
        "category": "Sexual Content",
        "isAdult": true
    },
    {
        "id": 534,
        "name": "Werewolf",
        "description": "Prominently features a character who is a werewolf.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 1760,
        "name": "Wilderness",
        "description": "Predominantly features a location with little to no human activity, such as a deserted island, a jungle, or a snowy mountain range.",
        "category": "Setting-Scene",
        "isAdult": false
    },
    {
        "id": 179,
        "name": "Witch",
        "description": "Prominently features a character who is a witch.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 145,
        "name": "Work",
        "description": "Centers around the activities of a certain occupation.",
        "category": "Setting-Scene",
        "isAdult": false
    },
    {
        "id": 231,
        "name": "Wrestling",
        "description": "Centers around the sport of wrestling.",
        "category": "Theme-Game-Sport",
        "isAdult": false
    },
    {
        "id": 394,
        "name": "Writing",
        "description": "Centers around the profession of writing books or novels.",
        "category": "Theme-Arts",
        "isAdult": false
    },
    {
        "id": 303,
        "name": "Wuxia",
        "description": "Chinese fiction concerning the adventures of martial artists in Ancient China.",
        "category": "Theme-Fantasy",
        "isAdult": false
    },
    {
        "id": 199,
        "name": "Yakuza",
        "description": "Centered around Japanese organised crime syndicates.",
        "category": "Theme-Other-Organisations",
        "isAdult": false
    },
    {
        "id": 163,
        "name": "Yandere",
        "description": "Prominently features a character who is obsessively in love with another, to the point of acting deranged or violent.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 233,
        "name": "Youkai",
        "description": "Prominently features supernatural creatures from Japanese folklore.",
        "category": "Theme-Fantasy",
        "isAdult": false
    },
    {
        "id": 76,
        "name": "Yuri",
        "description": "Prominently features romance between two females, not inherently sexual. Also known as Girls' Love.",
        "category": "Theme-Romance",
        "isAdult": false
    },
    {
        "id": 152,
        "name": "Zombie",
        "description": "Prominently features reanimated corpses which often prey on live humans and turn them into zombies.",
        "category": "Cast-Traits",
        "isAdult": false
    },
    {
        "id": 1446,
        "name": "Zoophilia",
        "description": "Features a character who has a sexual attraction for non-human animals.",
        "category": "Sexual Content",
        "isAdult": true
    }
]