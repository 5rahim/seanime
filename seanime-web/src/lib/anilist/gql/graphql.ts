/* eslint-disable */
import { TypedDocumentNode as DocumentNode } from "@graphql-typed-document-node/core"

export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
export type MakeEmpty<T extends { [key: string]: unknown }, K extends keyof T> = { [_ in K]?: never };
export type Incremental<T> = T | { [P in keyof T]?: P extends " $fragmentName" | "__typename" ? T[P] : never };
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
    ID: { input: string; output: string; }
    String: { input: string; output: string; }
    Boolean: { input: boolean; output: boolean; }
    Int: { input: number; output: number; }
    Float: { input: number; output: number; }
    /** ISO 3166-1 alpha-2 country code */
    CountryCode: { input: any; output: any; }
    /** 8 digit long date integer (YYYYMMDD). Unknown dates represented by 0. E.g. 2016: 20160000, May 1976: 19760500 */
    FuzzyDateInt: { input: any; output: any; }
    Json: { input: any; output: any; }
};

/** Notification for when a activity is liked */
export type ActivityLikeNotification = {
    /** The liked activity */
    activity?: Maybe<ActivityUnion>;
    /** The id of the activity which was liked */
    activityId: Scalars['Int']['output'];
    /** The notification context text */
    context?: Maybe<Scalars['String']['output']>;
    /** The time the notification was created at */
    createdAt?: Maybe<Scalars['Int']['output']>;
    /** The id of the Notification */
    id: Scalars['Int']['output'];
    /** The type of notification */
    type?: Maybe<NotificationType>;
    /** The user who liked the activity */
    user?: Maybe<User>;
    /** The id of the user who liked to the activity */
    userId: Scalars['Int']['output'];
};

/** Notification for when authenticated user is @ mentioned in activity or reply */
export type ActivityMentionNotification = {
    /** The liked activity */
    activity?: Maybe<ActivityUnion>;
    /** The id of the activity where mentioned */
    activityId: Scalars['Int']['output'];
    /** The notification context text */
    context?: Maybe<Scalars['String']['output']>;
    /** The time the notification was created at */
    createdAt?: Maybe<Scalars['Int']['output']>;
    /** The id of the Notification */
    id: Scalars['Int']['output'];
    /** The type of notification */
    type?: Maybe<NotificationType>;
    /** The user who mentioned the authenticated user */
    user?: Maybe<User>;
    /** The id of the user who mentioned the authenticated user */
    userId: Scalars['Int']['output'];
};

/** Notification for when a user is send an activity message */
export type ActivityMessageNotification = {
    /** The id of the activity message */
    activityId: Scalars['Int']['output'];
    /** The notification context text */
    context?: Maybe<Scalars['String']['output']>;
    /** The time the notification was created at */
    createdAt?: Maybe<Scalars['Int']['output']>;
    /** The id of the Notification */
    id: Scalars['Int']['output'];
    /** The message activity */
    message?: Maybe<MessageActivity>;
    /** The type of notification */
    type?: Maybe<NotificationType>;
    /** The user who sent the message */
    user?: Maybe<User>;
    /** The if of the user who send the message */
    userId: Scalars['Int']['output'];
};

/** Replay to an activity item */
export type ActivityReply = {
    /** The id of the parent activity */
    activityId?: Maybe<Scalars['Int']['output']>;
    /** The time the reply was created at */
    createdAt: Scalars['Int']['output'];
    /** The id of the reply */
    id: Scalars['Int']['output'];
    /** If the currently authenticated user liked the reply */
    isLiked?: Maybe<Scalars['Boolean']['output']>;
    /** The amount of likes the reply has */
    likeCount: Scalars['Int']['output'];
    /** The users who liked the reply */
    likes?: Maybe<Array<Maybe<User>>>;
    /** The reply text */
    text?: Maybe<Scalars['String']['output']>;
    /** The user who created reply */
    user?: Maybe<User>;
    /** The id of the replies creator */
    userId?: Maybe<Scalars['Int']['output']>;
};


/** Replay to an activity item */
export type ActivityReplyTextArgs = {
    asHtml?: InputMaybe<Scalars['Boolean']['input']>;
};

/** Notification for when a activity reply is liked */
export type ActivityReplyLikeNotification = {
    /** The liked activity */
    activity?: Maybe<ActivityUnion>;
    /** The id of the activity where the reply which was liked */
    activityId: Scalars['Int']['output'];
    /** The notification context text */
    context?: Maybe<Scalars['String']['output']>;
    /** The time the notification was created at */
    createdAt?: Maybe<Scalars['Int']['output']>;
    /** The id of the Notification */
    id: Scalars['Int']['output'];
    /** The type of notification */
    type?: Maybe<NotificationType>;
    /** The user who liked the activity reply */
    user?: Maybe<User>;
    /** The id of the user who liked to the activity reply */
    userId: Scalars['Int']['output'];
};

/** Notification for when a user replies to the authenticated users activity */
export type ActivityReplyNotification = {
    /** The liked activity */
    activity?: Maybe<ActivityUnion>;
    /** The id of the activity which was replied too */
    activityId: Scalars['Int']['output'];
    /** The notification context text */
    context?: Maybe<Scalars['String']['output']>;
    /** The time the notification was created at */
    createdAt?: Maybe<Scalars['Int']['output']>;
    /** The id of the Notification */
    id: Scalars['Int']['output'];
    /** The type of notification */
    type?: Maybe<NotificationType>;
    /** The user who replied to the activity */
    user?: Maybe<User>;
    /** The id of the user who replied to the activity */
    userId: Scalars['Int']['output'];
};

/** Notification for when a user replies to activity the authenticated user has replied to */
export type ActivityReplySubscribedNotification = {
    /** The liked activity */
    activity?: Maybe<ActivityUnion>;
    /** The id of the activity which was replied too */
    activityId: Scalars['Int']['output'];
    /** The notification context text */
    context?: Maybe<Scalars['String']['output']>;
    /** The time the notification was created at */
    createdAt?: Maybe<Scalars['Int']['output']>;
    /** The id of the Notification */
    id: Scalars['Int']['output'];
    /** The type of notification */
    type?: Maybe<NotificationType>;
    /** The user who replied to the activity */
    user?: Maybe<User>;
    /** The id of the user who replied to the activity */
    userId: Scalars['Int']['output'];
};

/** Activity sort enums */
export type ActivitySort =
    | 'ID'
    | 'ID_DESC'
    | 'PINNED';

/** Activity type enum. */
export type ActivityType =
/** A anime list update activity */
    | 'ANIME_LIST'
    /** A manga list update activity */
    | 'MANGA_LIST'
    /** Anime & Manga list update, only used in query arguments */
    | 'MEDIA_LIST'
    /** A text message activity sent to another user */
    | 'MESSAGE'
    /** A text activity */
    | 'TEXT';

/** Activity union type */
export type ActivityUnion = ListActivity | MessageActivity | TextActivity;

/** Notification for when an episode of anime airs */
export type AiringNotification = {
    /** The id of the aired anime */
    animeId: Scalars['Int']['output'];
    /** The notification context text */
    contexts?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
    /** The time the notification was created at */
    createdAt?: Maybe<Scalars['Int']['output']>;
    /** The episode number that just aired */
    episode: Scalars['Int']['output'];
    /** The id of the Notification */
    id: Scalars['Int']['output'];
    /** The associated media of the airing schedule */
    media?: Maybe<Media>;
    /** The type of notification */
    type?: Maybe<NotificationType>;
};

/** Score & Watcher stats for airing anime by episode and mid-week */
export type AiringProgression = {
    /** The episode the stats were recorded at. .5 is the mid point between 2 episodes airing dates. */
    episode?: Maybe<Scalars['Float']['output']>;
    /** The average score for the media */
    score?: Maybe<Scalars['Float']['output']>;
    /** The amount of users watching the anime */
    watching?: Maybe<Scalars['Int']['output']>;
};

/** Media Airing Schedule. NOTE: We only aim to guarantee that FUTURE airing data is present and accurate. */
export type AiringSchedule = {
    /** The time the episode airs at */
    airingAt: Scalars['Int']['output'];
    /** The airing episode number */
    episode: Scalars['Int']['output'];
    /** The id of the airing schedule item */
    id: Scalars['Int']['output'];
    /** The associate media of the airing episode */
    media?: Maybe<Media>;
    /** The associate media id of the airing episode */
    mediaId: Scalars['Int']['output'];
    /** Seconds until episode starts airing */
    timeUntilAiring: Scalars['Int']['output'];
};

export type AiringScheduleConnection = {
    edges?: Maybe<Array<Maybe<AiringScheduleEdge>>>;
    nodes?: Maybe<Array<Maybe<AiringSchedule>>>;
    /** The pagination information */
    pageInfo?: Maybe<PageInfo>;
};

/** AiringSchedule connection edge */
export type AiringScheduleEdge = {
    /** The id of the connection */
    id?: Maybe<Scalars['Int']['output']>;
    node?: Maybe<AiringSchedule>;
};

export type AiringScheduleInput = {
    airingAt?: InputMaybe<Scalars['Int']['input']>;
    episode?: InputMaybe<Scalars['Int']['input']>;
    timeUntilAiring?: InputMaybe<Scalars['Int']['input']>;
};

/** Airing schedule sort enums */
export type AiringSort =
    | 'EPISODE'
    | 'EPISODE_DESC'
    | 'ID'
    | 'ID_DESC'
    | 'MEDIA_ID'
    | 'MEDIA_ID_DESC'
    | 'TIME'
    | 'TIME_DESC';

export type AniChartHighlightInput = {
    highlight?: InputMaybe<Scalars['String']['input']>;
    mediaId?: InputMaybe<Scalars['Int']['input']>;
};

export type AniChartUser = {
    highlights?: Maybe<Scalars['Json']['output']>;
    settings?: Maybe<Scalars['Json']['output']>;
    user?: Maybe<User>;
};

/** A character that features in an anime or manga */
export type Character = {
    /** The character's age. Note this is a string, not an int, it may contain further text and additional ages. */
    age?: Maybe<Scalars['String']['output']>;
    /** The characters blood type */
    bloodType?: Maybe<Scalars['String']['output']>;
    /** The character's birth date */
    dateOfBirth?: Maybe<FuzzyDate>;
    /** A general description of the character */
    description?: Maybe<Scalars['String']['output']>;
    /** The amount of user's who have favourited the character */
    favourites?: Maybe<Scalars['Int']['output']>;
    /** The character's gender. Usually Male, Female, or Non-binary but can be any string. */
    gender?: Maybe<Scalars['String']['output']>;
    /** The id of the character */
    id: Scalars['Int']['output'];
    /** Character images */
    image?: Maybe<CharacterImage>;
    /** If the character is marked as favourite by the currently authenticated user */
    isFavourite: Scalars['Boolean']['output'];
    /** If the character is blocked from being added to favourites */
    isFavouriteBlocked: Scalars['Boolean']['output'];
    /** Media that includes the character */
    media?: Maybe<MediaConnection>;
    /** Notes for site moderators */
    modNotes?: Maybe<Scalars['String']['output']>;
    /** The names of the character */
    name?: Maybe<CharacterName>;
    /** The url for the character page on the AniList website */
    siteUrl?: Maybe<Scalars['String']['output']>;
    /** @deprecated No data available */
    updatedAt?: Maybe<Scalars['Int']['output']>;
};


/** A character that features in an anime or manga */
export type CharacterDescriptionArgs = {
    asHtml?: InputMaybe<Scalars['Boolean']['input']>;
};


/** A character that features in an anime or manga */
export type CharacterMediaArgs = {
    onList?: InputMaybe<Scalars['Boolean']['input']>;
    page?: InputMaybe<Scalars['Int']['input']>;
    perPage?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<MediaSort>>>;
    type?: InputMaybe<MediaType>;
};

export type CharacterConnection = {
    edges?: Maybe<Array<Maybe<CharacterEdge>>>;
    nodes?: Maybe<Array<Maybe<Character>>>;
    /** The pagination information */
    pageInfo?: Maybe<PageInfo>;
};

/** Character connection edge */
export type CharacterEdge = {
    /** The order the character should be displayed from the users favourites */
    favouriteOrder?: Maybe<Scalars['Int']['output']>;
    /** The id of the connection */
    id?: Maybe<Scalars['Int']['output']>;
    /** The media the character is in */
    media?: Maybe<Array<Maybe<Media>>>;
    /** Media specific character name */
    name?: Maybe<Scalars['String']['output']>;
    node?: Maybe<Character>;
    /** The characters role in the media */
    role?: Maybe<CharacterRole>;
    /** The voice actors of the character with role date */
    voiceActorRoles?: Maybe<Array<Maybe<StaffRoleType>>>;
    /** The voice actors of the character */
    voiceActors?: Maybe<Array<Maybe<Staff>>>;
};


/** Character connection edge */
export type CharacterEdgeVoiceActorRolesArgs = {
    language?: InputMaybe<StaffLanguage>;
    sort?: InputMaybe<Array<InputMaybe<StaffSort>>>;
};


/** Character connection edge */
export type CharacterEdgeVoiceActorsArgs = {
    language?: InputMaybe<StaffLanguage>;
    sort?: InputMaybe<Array<InputMaybe<StaffSort>>>;
};

export type CharacterImage = {
    /** The character's image of media at its largest size */
    large?: Maybe<Scalars['String']['output']>;
    /** The character's image of media at medium size */
    medium?: Maybe<Scalars['String']['output']>;
};

/** The names of the character */
export type CharacterName = {
    /** Other names the character might be referred to as */
    alternative?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
    /** Other names the character might be referred to as but are spoilers */
    alternativeSpoiler?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
    /** The character's given name */
    first?: Maybe<Scalars['String']['output']>;
    /** The character's first and last name */
    full?: Maybe<Scalars['String']['output']>;
    /** The character's surname */
    last?: Maybe<Scalars['String']['output']>;
    /** The character's middle name */
    middle?: Maybe<Scalars['String']['output']>;
    /** The character's full name in their native language */
    native?: Maybe<Scalars['String']['output']>;
    /** The currently authenticated users preferred name language. Default romaji for non-authenticated */
    userPreferred?: Maybe<Scalars['String']['output']>;
};

/** The names of the character */
export type CharacterNameInput = {
    /** Other names the character might be referred by */
    alternative?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    /** Other names the character might be referred to as but are spoilers */
    alternativeSpoiler?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    /** The character's given name */
    first?: InputMaybe<Scalars['String']['input']>;
    /** The character's surname */
    last?: InputMaybe<Scalars['String']['input']>;
    /** The character's middle name */
    middle?: InputMaybe<Scalars['String']['input']>;
    /** The character's full name in their native language */
    native?: InputMaybe<Scalars['String']['input']>;
};

/** The role the character plays in the media */
export type CharacterRole =
/** A background character in the media */
    | 'BACKGROUND'
    /** A primary character role in the media */
    | 'MAIN'
    /** A supporting character role in the media */
    | 'SUPPORTING';

/** Character sort enums */
export type CharacterSort =
    | 'FAVOURITES'
    | 'FAVOURITES_DESC'
    | 'ID'
    | 'ID_DESC'
    /** Order manually decided by moderators */
    | 'RELEVANCE'
    | 'ROLE'
    | 'ROLE_DESC'
    | 'SEARCH_MATCH';

/** A submission for a character that features in an anime or manga */
export type CharacterSubmission = {
    /** Data Mod assigned to handle the submission */
    assignee?: Maybe<User>;
    /** Character that the submission is referencing */
    character?: Maybe<Character>;
    createdAt?: Maybe<Scalars['Int']['output']>;
    /** The id of the submission */
    id: Scalars['Int']['output'];
    /** Whether the submission is locked */
    locked?: Maybe<Scalars['Boolean']['output']>;
    /** Inner details of submission status */
    notes?: Maybe<Scalars['String']['output']>;
    source?: Maybe<Scalars['String']['output']>;
    /** Status of the submission */
    status?: Maybe<SubmissionStatus>;
    /** The character submission changes */
    submission?: Maybe<Character>;
    /** Submitter for the submission */
    submitter?: Maybe<User>;
};

export type CharacterSubmissionConnection = {
    edges?: Maybe<Array<Maybe<CharacterSubmissionEdge>>>;
    nodes?: Maybe<Array<Maybe<CharacterSubmission>>>;
    /** The pagination information */
    pageInfo?: Maybe<PageInfo>;
};

/** CharacterSubmission connection edge */
export type CharacterSubmissionEdge = {
    node?: Maybe<CharacterSubmission>;
    /** The characters role in the media */
    role?: Maybe<CharacterRole>;
    /** The submitted voice actors of the character */
    submittedVoiceActors?: Maybe<Array<Maybe<StaffSubmission>>>;
    /** The voice actors of the character */
    voiceActors?: Maybe<Array<Maybe<Staff>>>;
};

/** Deleted data type */
export type Deleted = {
    /** If an item has been successfully deleted */
    deleted?: Maybe<Scalars['Boolean']['output']>;
};

export type ExternalLinkMediaType =
    | 'ANIME'
    | 'MANGA'
    | 'STAFF';

export type ExternalLinkType =
    | 'INFO'
    | 'SOCIAL'
    | 'STREAMING';

/** User's favourite anime, manga, characters, staff & studios */
export type Favourites = {
    /** Favourite anime */
    anime?: Maybe<MediaConnection>;
    /** Favourite characters */
    characters?: Maybe<CharacterConnection>;
    /** Favourite manga */
    manga?: Maybe<MediaConnection>;
    /** Favourite staff */
    staff?: Maybe<StaffConnection>;
    /** Favourite studios */
    studios?: Maybe<StudioConnection>;
};


/** User's favourite anime, manga, characters, staff & studios */
export type FavouritesAnimeArgs = {
    page?: InputMaybe<Scalars['Int']['input']>;
    perPage?: InputMaybe<Scalars['Int']['input']>;
};


/** User's favourite anime, manga, characters, staff & studios */
export type FavouritesCharactersArgs = {
    page?: InputMaybe<Scalars['Int']['input']>;
    perPage?: InputMaybe<Scalars['Int']['input']>;
};


/** User's favourite anime, manga, characters, staff & studios */
export type FavouritesMangaArgs = {
    page?: InputMaybe<Scalars['Int']['input']>;
    perPage?: InputMaybe<Scalars['Int']['input']>;
};


/** User's favourite anime, manga, characters, staff & studios */
export type FavouritesStaffArgs = {
    page?: InputMaybe<Scalars['Int']['input']>;
    perPage?: InputMaybe<Scalars['Int']['input']>;
};


/** User's favourite anime, manga, characters, staff & studios */
export type FavouritesStudiosArgs = {
    page?: InputMaybe<Scalars['Int']['input']>;
    perPage?: InputMaybe<Scalars['Int']['input']>;
};

/** Notification for when the authenticated user is followed by another user */
export type FollowingNotification = {
    /** The notification context text */
    context?: Maybe<Scalars['String']['output']>;
    /** The time the notification was created at */
    createdAt?: Maybe<Scalars['Int']['output']>;
    /** The id of the Notification */
    id: Scalars['Int']['output'];
    /** The type of notification */
    type?: Maybe<NotificationType>;
    /** The liked activity */
    user?: Maybe<User>;
    /** The id of the user who followed the authenticated user */
    userId: Scalars['Int']['output'];
};

/** User's format statistics */
export type FormatStats = {
    amount?: Maybe<Scalars['Int']['output']>;
    format?: Maybe<MediaFormat>;
};

/** Date object that allows for incomplete date values (fuzzy) */
export type FuzzyDate = {
    /** Numeric Day (24) */
    day?: Maybe<Scalars['Int']['output']>;
    /** Numeric Month (3) */
    month?: Maybe<Scalars['Int']['output']>;
    /** Numeric Year (2017) */
    year?: Maybe<Scalars['Int']['output']>;
};

/** Date object that allows for incomplete date values (fuzzy) */
export type FuzzyDateInput = {
    /** Numeric Day (24) */
    day?: InputMaybe<Scalars['Int']['input']>;
    /** Numeric Month (3) */
    month?: InputMaybe<Scalars['Int']['input']>;
    /** Numeric Year (2017) */
    year?: InputMaybe<Scalars['Int']['input']>;
};

/** User's genre statistics */
export type GenreStats = {
    amount?: Maybe<Scalars['Int']['output']>;
    genre?: Maybe<Scalars['String']['output']>;
    meanScore?: Maybe<Scalars['Int']['output']>;
    /** The amount of time in minutes the genre has been watched by the user */
    timeWatched?: Maybe<Scalars['Int']['output']>;
};

/** Page of data (Used for internal use only) */
export type InternalPage = {
    activities?: Maybe<Array<Maybe<ActivityUnion>>>;
    activityReplies?: Maybe<Array<Maybe<ActivityReply>>>;
    airingSchedules?: Maybe<Array<Maybe<AiringSchedule>>>;
    characterSubmissions?: Maybe<Array<Maybe<CharacterSubmission>>>;
    characters?: Maybe<Array<Maybe<Character>>>;
    followers?: Maybe<Array<Maybe<User>>>;
    following?: Maybe<Array<Maybe<User>>>;
    likes?: Maybe<Array<Maybe<User>>>;
    media?: Maybe<Array<Maybe<Media>>>;
    mediaList?: Maybe<Array<Maybe<MediaList>>>;
    mediaSubmissions?: Maybe<Array<Maybe<MediaSubmission>>>;
    mediaTrends?: Maybe<Array<Maybe<MediaTrend>>>;
    modActions?: Maybe<Array<Maybe<ModAction>>>;
    notifications?: Maybe<Array<Maybe<NotificationUnion>>>;
    /** The pagination information */
    pageInfo?: Maybe<PageInfo>;
    recommendations?: Maybe<Array<Maybe<Recommendation>>>;
    reports?: Maybe<Array<Maybe<Report>>>;
    reviews?: Maybe<Array<Maybe<Review>>>;
    revisionHistory?: Maybe<Array<Maybe<RevisionHistory>>>;
    staff?: Maybe<Array<Maybe<Staff>>>;
    staffSubmissions?: Maybe<Array<Maybe<StaffSubmission>>>;
    studios?: Maybe<Array<Maybe<Studio>>>;
    threadComments?: Maybe<Array<Maybe<ThreadComment>>>;
    threads?: Maybe<Array<Maybe<Thread>>>;
    userBlockSearch?: Maybe<Array<Maybe<User>>>;
    users?: Maybe<Array<Maybe<User>>>;
};


/** Page of data (Used for internal use only) */
export type InternalPageActivitiesArgs = {
    createdAt?: InputMaybe<Scalars['Int']['input']>;
    createdAt_greater?: InputMaybe<Scalars['Int']['input']>;
    createdAt_lesser?: InputMaybe<Scalars['Int']['input']>;
    hasReplies?: InputMaybe<Scalars['Boolean']['input']>;
    hasRepliesOrTypeText?: InputMaybe<Scalars['Boolean']['input']>;
    id?: InputMaybe<Scalars['Int']['input']>;
    id_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    id_not?: InputMaybe<Scalars['Int']['input']>;
    id_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    isFollowing?: InputMaybe<Scalars['Boolean']['input']>;
    mediaId?: InputMaybe<Scalars['Int']['input']>;
    mediaId_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    mediaId_not?: InputMaybe<Scalars['Int']['input']>;
    mediaId_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    messengerId?: InputMaybe<Scalars['Int']['input']>;
    messengerId_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    messengerId_not?: InputMaybe<Scalars['Int']['input']>;
    messengerId_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    sort?: InputMaybe<Array<InputMaybe<ActivitySort>>>;
    type?: InputMaybe<ActivityType>;
    type_in?: InputMaybe<Array<InputMaybe<ActivityType>>>;
    type_not?: InputMaybe<ActivityType>;
    type_not_in?: InputMaybe<Array<InputMaybe<ActivityType>>>;
    userId?: InputMaybe<Scalars['Int']['input']>;
    userId_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    userId_not?: InputMaybe<Scalars['Int']['input']>;
    userId_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
};


/** Page of data (Used for internal use only) */
export type InternalPageActivityRepliesArgs = {
    activityId?: InputMaybe<Scalars['Int']['input']>;
    id?: InputMaybe<Scalars['Int']['input']>;
};


/** Page of data (Used for internal use only) */
export type InternalPageAiringSchedulesArgs = {
    airingAt?: InputMaybe<Scalars['Int']['input']>;
    airingAt_greater?: InputMaybe<Scalars['Int']['input']>;
    airingAt_lesser?: InputMaybe<Scalars['Int']['input']>;
    episode?: InputMaybe<Scalars['Int']['input']>;
    episode_greater?: InputMaybe<Scalars['Int']['input']>;
    episode_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    episode_lesser?: InputMaybe<Scalars['Int']['input']>;
    episode_not?: InputMaybe<Scalars['Int']['input']>;
    episode_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    id?: InputMaybe<Scalars['Int']['input']>;
    id_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    id_not?: InputMaybe<Scalars['Int']['input']>;
    id_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    mediaId?: InputMaybe<Scalars['Int']['input']>;
    mediaId_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    mediaId_not?: InputMaybe<Scalars['Int']['input']>;
    mediaId_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    notYetAired?: InputMaybe<Scalars['Boolean']['input']>;
    sort?: InputMaybe<Array<InputMaybe<AiringSort>>>;
};


/** Page of data (Used for internal use only) */
export type InternalPageCharacterSubmissionsArgs = {
    assigneeId?: InputMaybe<Scalars['Int']['input']>;
    characterId?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<SubmissionSort>>>;
    status?: InputMaybe<SubmissionStatus>;
    userId?: InputMaybe<Scalars['Int']['input']>;
};


/** Page of data (Used for internal use only) */
export type InternalPageCharactersArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
    id_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    id_not?: InputMaybe<Scalars['Int']['input']>;
    id_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    isBirthday?: InputMaybe<Scalars['Boolean']['input']>;
    search?: InputMaybe<Scalars['String']['input']>;
    sort?: InputMaybe<Array<InputMaybe<CharacterSort>>>;
};


/** Page of data (Used for internal use only) */
export type InternalPageFollowersArgs = {
    sort?: InputMaybe<Array<InputMaybe<UserSort>>>;
    userId: Scalars['Int']['input'];
};


/** Page of data (Used for internal use only) */
export type InternalPageFollowingArgs = {
    sort?: InputMaybe<Array<InputMaybe<UserSort>>>;
    userId: Scalars['Int']['input'];
};


/** Page of data (Used for internal use only) */
export type InternalPageLikesArgs = {
    likeableId?: InputMaybe<Scalars['Int']['input']>;
    type?: InputMaybe<LikeableType>;
};


/** Page of data (Used for internal use only) */
export type InternalPageMediaArgs = {
    averageScore?: InputMaybe<Scalars['Int']['input']>;
    averageScore_greater?: InputMaybe<Scalars['Int']['input']>;
    averageScore_lesser?: InputMaybe<Scalars['Int']['input']>;
    averageScore_not?: InputMaybe<Scalars['Int']['input']>;
    chapters?: InputMaybe<Scalars['Int']['input']>;
    chapters_greater?: InputMaybe<Scalars['Int']['input']>;
    chapters_lesser?: InputMaybe<Scalars['Int']['input']>;
    countryOfOrigin?: InputMaybe<Scalars['CountryCode']['input']>;
    duration?: InputMaybe<Scalars['Int']['input']>;
    duration_greater?: InputMaybe<Scalars['Int']['input']>;
    duration_lesser?: InputMaybe<Scalars['Int']['input']>;
    endDate?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    endDate_greater?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    endDate_lesser?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    endDate_like?: InputMaybe<Scalars['String']['input']>;
    episodes?: InputMaybe<Scalars['Int']['input']>;
    episodes_greater?: InputMaybe<Scalars['Int']['input']>;
    episodes_lesser?: InputMaybe<Scalars['Int']['input']>;
    format?: InputMaybe<MediaFormat>;
    format_in?: InputMaybe<Array<InputMaybe<MediaFormat>>>;
    format_not?: InputMaybe<MediaFormat>;
    format_not_in?: InputMaybe<Array<InputMaybe<MediaFormat>>>;
    genre?: InputMaybe<Scalars['String']['input']>;
    genre_in?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    genre_not_in?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    id?: InputMaybe<Scalars['Int']['input']>;
    idMal?: InputMaybe<Scalars['Int']['input']>;
    idMal_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    idMal_not?: InputMaybe<Scalars['Int']['input']>;
    idMal_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    id_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    id_not?: InputMaybe<Scalars['Int']['input']>;
    id_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    isAdult?: InputMaybe<Scalars['Boolean']['input']>;
    isLicensed?: InputMaybe<Scalars['Boolean']['input']>;
    licensedBy?: InputMaybe<Scalars['String']['input']>;
    licensedById?: InputMaybe<Scalars['Int']['input']>;
    licensedById_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    licensedBy_in?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    minimumTagRank?: InputMaybe<Scalars['Int']['input']>;
    onList?: InputMaybe<Scalars['Boolean']['input']>;
    popularity?: InputMaybe<Scalars['Int']['input']>;
    popularity_greater?: InputMaybe<Scalars['Int']['input']>;
    popularity_lesser?: InputMaybe<Scalars['Int']['input']>;
    popularity_not?: InputMaybe<Scalars['Int']['input']>;
    search?: InputMaybe<Scalars['String']['input']>;
    season?: InputMaybe<MediaSeason>;
    seasonYear?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<MediaSort>>>;
    source?: InputMaybe<MediaSource>;
    source_in?: InputMaybe<Array<InputMaybe<MediaSource>>>;
    startDate?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    startDate_greater?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    startDate_lesser?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    startDate_like?: InputMaybe<Scalars['String']['input']>;
    status?: InputMaybe<MediaStatus>;
    status_in?: InputMaybe<Array<InputMaybe<MediaStatus>>>;
    status_not?: InputMaybe<MediaStatus>;
    status_not_in?: InputMaybe<Array<InputMaybe<MediaStatus>>>;
    tag?: InputMaybe<Scalars['String']['input']>;
    tagCategory?: InputMaybe<Scalars['String']['input']>;
    tagCategory_in?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    tagCategory_not_in?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    tag_in?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    tag_not_in?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    type?: InputMaybe<MediaType>;
    volumes?: InputMaybe<Scalars['Int']['input']>;
    volumes_greater?: InputMaybe<Scalars['Int']['input']>;
    volumes_lesser?: InputMaybe<Scalars['Int']['input']>;
};


/** Page of data (Used for internal use only) */
export type InternalPageMediaListArgs = {
    compareWithAuthList?: InputMaybe<Scalars['Boolean']['input']>;
    completedAt?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    completedAt_greater?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    completedAt_lesser?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    completedAt_like?: InputMaybe<Scalars['String']['input']>;
    id?: InputMaybe<Scalars['Int']['input']>;
    isFollowing?: InputMaybe<Scalars['Boolean']['input']>;
    mediaId?: InputMaybe<Scalars['Int']['input']>;
    mediaId_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    mediaId_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    notes?: InputMaybe<Scalars['String']['input']>;
    notes_like?: InputMaybe<Scalars['String']['input']>;
    sort?: InputMaybe<Array<InputMaybe<MediaListSort>>>;
    startedAt?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    startedAt_greater?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    startedAt_lesser?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    startedAt_like?: InputMaybe<Scalars['String']['input']>;
    status?: InputMaybe<MediaListStatus>;
    status_in?: InputMaybe<Array<InputMaybe<MediaListStatus>>>;
    status_not?: InputMaybe<MediaListStatus>;
    status_not_in?: InputMaybe<Array<InputMaybe<MediaListStatus>>>;
    type?: InputMaybe<MediaType>;
    userId?: InputMaybe<Scalars['Int']['input']>;
    userId_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    userName?: InputMaybe<Scalars['String']['input']>;
};


/** Page of data (Used for internal use only) */
export type InternalPageMediaSubmissionsArgs = {
    assigneeId?: InputMaybe<Scalars['Int']['input']>;
    mediaId?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<SubmissionSort>>>;
    status?: InputMaybe<SubmissionStatus>;
    submissionId?: InputMaybe<Scalars['Int']['input']>;
    type?: InputMaybe<MediaType>;
    userId?: InputMaybe<Scalars['Int']['input']>;
};


/** Page of data (Used for internal use only) */
export type InternalPageMediaTrendsArgs = {
    averageScore?: InputMaybe<Scalars['Int']['input']>;
    averageScore_greater?: InputMaybe<Scalars['Int']['input']>;
    averageScore_lesser?: InputMaybe<Scalars['Int']['input']>;
    averageScore_not?: InputMaybe<Scalars['Int']['input']>;
    date?: InputMaybe<Scalars['Int']['input']>;
    date_greater?: InputMaybe<Scalars['Int']['input']>;
    date_lesser?: InputMaybe<Scalars['Int']['input']>;
    episode?: InputMaybe<Scalars['Int']['input']>;
    episode_greater?: InputMaybe<Scalars['Int']['input']>;
    episode_lesser?: InputMaybe<Scalars['Int']['input']>;
    episode_not?: InputMaybe<Scalars['Int']['input']>;
    mediaId?: InputMaybe<Scalars['Int']['input']>;
    mediaId_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    mediaId_not?: InputMaybe<Scalars['Int']['input']>;
    mediaId_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    popularity?: InputMaybe<Scalars['Int']['input']>;
    popularity_greater?: InputMaybe<Scalars['Int']['input']>;
    popularity_lesser?: InputMaybe<Scalars['Int']['input']>;
    popularity_not?: InputMaybe<Scalars['Int']['input']>;
    releasing?: InputMaybe<Scalars['Boolean']['input']>;
    sort?: InputMaybe<Array<InputMaybe<MediaTrendSort>>>;
    trending?: InputMaybe<Scalars['Int']['input']>;
    trending_greater?: InputMaybe<Scalars['Int']['input']>;
    trending_lesser?: InputMaybe<Scalars['Int']['input']>;
    trending_not?: InputMaybe<Scalars['Int']['input']>;
};


/** Page of data (Used for internal use only) */
export type InternalPageModActionsArgs = {
    modId?: InputMaybe<Scalars['Int']['input']>;
    userId?: InputMaybe<Scalars['Int']['input']>;
};


/** Page of data (Used for internal use only) */
export type InternalPageNotificationsArgs = {
    resetNotificationCount?: InputMaybe<Scalars['Boolean']['input']>;
    type?: InputMaybe<NotificationType>;
    type_in?: InputMaybe<Array<InputMaybe<NotificationType>>>;
};


/** Page of data (Used for internal use only) */
export type InternalPageRecommendationsArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
    mediaId?: InputMaybe<Scalars['Int']['input']>;
    mediaRecommendationId?: InputMaybe<Scalars['Int']['input']>;
    onList?: InputMaybe<Scalars['Boolean']['input']>;
    rating?: InputMaybe<Scalars['Int']['input']>;
    rating_greater?: InputMaybe<Scalars['Int']['input']>;
    rating_lesser?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<RecommendationSort>>>;
    userId?: InputMaybe<Scalars['Int']['input']>;
};


/** Page of data (Used for internal use only) */
export type InternalPageReportsArgs = {
    reportedId?: InputMaybe<Scalars['Int']['input']>;
    reporterId?: InputMaybe<Scalars['Int']['input']>;
};


/** Page of data (Used for internal use only) */
export type InternalPageReviewsArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
    mediaId?: InputMaybe<Scalars['Int']['input']>;
    mediaType?: InputMaybe<MediaType>;
    sort?: InputMaybe<Array<InputMaybe<ReviewSort>>>;
    userId?: InputMaybe<Scalars['Int']['input']>;
};


/** Page of data (Used for internal use only) */
export type InternalPageRevisionHistoryArgs = {
    characterId?: InputMaybe<Scalars['Int']['input']>;
    mediaId?: InputMaybe<Scalars['Int']['input']>;
    staffId?: InputMaybe<Scalars['Int']['input']>;
    studioId?: InputMaybe<Scalars['Int']['input']>;
    userId?: InputMaybe<Scalars['Int']['input']>;
};


/** Page of data (Used for internal use only) */
export type InternalPageStaffArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
    id_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    id_not?: InputMaybe<Scalars['Int']['input']>;
    id_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    isBirthday?: InputMaybe<Scalars['Boolean']['input']>;
    search?: InputMaybe<Scalars['String']['input']>;
    sort?: InputMaybe<Array<InputMaybe<StaffSort>>>;
};


/** Page of data (Used for internal use only) */
export type InternalPageStaffSubmissionsArgs = {
    assigneeId?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<SubmissionSort>>>;
    staffId?: InputMaybe<Scalars['Int']['input']>;
    status?: InputMaybe<SubmissionStatus>;
    userId?: InputMaybe<Scalars['Int']['input']>;
};


/** Page of data (Used for internal use only) */
export type InternalPageStudiosArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
    id_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    id_not?: InputMaybe<Scalars['Int']['input']>;
    id_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    search?: InputMaybe<Scalars['String']['input']>;
    sort?: InputMaybe<Array<InputMaybe<StudioSort>>>;
};


/** Page of data (Used for internal use only) */
export type InternalPageThreadCommentsArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<ThreadCommentSort>>>;
    threadId?: InputMaybe<Scalars['Int']['input']>;
    userId?: InputMaybe<Scalars['Int']['input']>;
};


/** Page of data (Used for internal use only) */
export type InternalPageThreadsArgs = {
    categoryId?: InputMaybe<Scalars['Int']['input']>;
    id?: InputMaybe<Scalars['Int']['input']>;
    id_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    mediaCategoryId?: InputMaybe<Scalars['Int']['input']>;
    replyUserId?: InputMaybe<Scalars['Int']['input']>;
    search?: InputMaybe<Scalars['String']['input']>;
    sort?: InputMaybe<Array<InputMaybe<ThreadSort>>>;
    subscribed?: InputMaybe<Scalars['Boolean']['input']>;
    userId?: InputMaybe<Scalars['Int']['input']>;
};


/** Page of data (Used for internal use only) */
export type InternalPageUserBlockSearchArgs = {
    search?: InputMaybe<Scalars['String']['input']>;
};


/** Page of data (Used for internal use only) */
export type InternalPageUsersArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
    isModerator?: InputMaybe<Scalars['Boolean']['input']>;
    name?: InputMaybe<Scalars['String']['input']>;
    search?: InputMaybe<Scalars['String']['input']>;
    sort?: InputMaybe<Array<InputMaybe<UserSort>>>;
};

/** Types that can be liked */
export type LikeableType =
    | 'ACTIVITY'
    | 'ACTIVITY_REPLY'
    | 'THREAD'
    | 'THREAD_COMMENT';

/** Likeable union type */
export type LikeableUnion = ActivityReply | ListActivity | MessageActivity | TextActivity | Thread | ThreadComment;

/** User list activity (anime & manga updates) */
export type ListActivity = {
    /** The time the activity was created at */
    createdAt: Scalars['Int']['output'];
    /** The id of the activity */
    id: Scalars['Int']['output'];
    /** If the currently authenticated user liked the activity */
    isLiked?: Maybe<Scalars['Boolean']['output']>;
    /** If the activity is locked and can receive replies */
    isLocked?: Maybe<Scalars['Boolean']['output']>;
    /** If the activity is pinned to the top of the users activity feed */
    isPinned?: Maybe<Scalars['Boolean']['output']>;
    /** If the currently authenticated user is subscribed to the activity */
    isSubscribed?: Maybe<Scalars['Boolean']['output']>;
    /** The amount of likes the activity has */
    likeCount: Scalars['Int']['output'];
    /** The users who liked the activity */
    likes?: Maybe<Array<Maybe<User>>>;
    /** The associated media to the activity update */
    media?: Maybe<Media>;
    /** The list progress made */
    progress?: Maybe<Scalars['String']['output']>;
    /** The written replies to the activity */
    replies?: Maybe<Array<Maybe<ActivityReply>>>;
    /** The number of activity replies */
    replyCount: Scalars['Int']['output'];
    /** The url for the activity page on the AniList website */
    siteUrl?: Maybe<Scalars['String']['output']>;
    /** The list item's textual status */
    status?: Maybe<Scalars['String']['output']>;
    /** The type of activity */
    type?: Maybe<ActivityType>;
    /** The owner of the activity */
    user?: Maybe<User>;
    /** The user id of the activity's creator */
    userId?: Maybe<Scalars['Int']['output']>;
};

export type ListActivityOption = {
    disabled?: Maybe<Scalars['Boolean']['output']>;
    type?: Maybe<MediaListStatus>;
};

export type ListActivityOptionInput = {
    disabled?: InputMaybe<Scalars['Boolean']['input']>;
    type?: InputMaybe<MediaListStatus>;
};

/** User's list score statistics */
export type ListScoreStats = {
    meanScore?: Maybe<Scalars['Int']['output']>;
    standardDeviation?: Maybe<Scalars['Int']['output']>;
};

/** Anime or Manga */
export type Media = {
    /** The media's entire airing schedule */
    airingSchedule?: Maybe<AiringScheduleConnection>;
    /** If the media should have forum thread automatically created for it on airing episode release */
    autoCreateForumThread?: Maybe<Scalars['Boolean']['output']>;
    /** A weighted average score of all the user's scores of the media */
    averageScore?: Maybe<Scalars['Int']['output']>;
    /** The banner image of the media */
    bannerImage?: Maybe<Scalars['String']['output']>;
    /** The amount of chapters the manga has when complete */
    chapters?: Maybe<Scalars['Int']['output']>;
    /** The characters in the media */
    characters?: Maybe<CharacterConnection>;
    /** Where the media was created. (ISO 3166-1 alpha-2) */
    countryOfOrigin?: Maybe<Scalars['CountryCode']['output']>;
    /** The cover images of the media */
    coverImage?: Maybe<MediaCoverImage>;
    /** Short description of the media's story and characters */
    description?: Maybe<Scalars['String']['output']>;
    /** The general length of each anime episode in minutes */
    duration?: Maybe<Scalars['Int']['output']>;
    /** The last official release date of the media */
    endDate?: Maybe<FuzzyDate>;
    /** The amount of episodes the anime has when complete */
    episodes?: Maybe<Scalars['Int']['output']>;
    /** External links to another site related to the media */
    externalLinks?: Maybe<Array<Maybe<MediaExternalLink>>>;
    /** The amount of user's who have favourited the media */
    favourites?: Maybe<Scalars['Int']['output']>;
    /** The format the media was released in */
    format?: Maybe<MediaFormat>;
    /** The genres of the media */
    genres?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
    /** Official Twitter hashtags for the media */
    hashtag?: Maybe<Scalars['String']['output']>;
    /** The id of the media */
    id: Scalars['Int']['output'];
    /** The mal id of the media */
    idMal?: Maybe<Scalars['Int']['output']>;
    /** If the media is intended only for 18+ adult audiences */
    isAdult?: Maybe<Scalars['Boolean']['output']>;
    /** If the media is marked as favourite by the current authenticated user */
    isFavourite: Scalars['Boolean']['output'];
    /** If the media is blocked from being added to favourites */
    isFavouriteBlocked: Scalars['Boolean']['output'];
    /** If the media is officially licensed or a self-published doujin release */
    isLicensed?: Maybe<Scalars['Boolean']['output']>;
    /** Locked media may not be added to lists our favorited. This may be due to the entry pending for deletion or other reasons. */
    isLocked?: Maybe<Scalars['Boolean']['output']>;
    /** If the media is blocked from being recommended to/from */
    isRecommendationBlocked?: Maybe<Scalars['Boolean']['output']>;
    /** If the media is blocked from being reviewed */
    isReviewBlocked?: Maybe<Scalars['Boolean']['output']>;
    /** Mean score of all the user's scores of the media */
    meanScore?: Maybe<Scalars['Int']['output']>;
    /** The authenticated user's media list entry for the media */
    mediaListEntry?: Maybe<MediaList>;
    /** Notes for site moderators */
    modNotes?: Maybe<Scalars['String']['output']>;
    /** The media's next episode airing schedule */
    nextAiringEpisode?: Maybe<AiringSchedule>;
    /** The number of users with the media on their list */
    popularity?: Maybe<Scalars['Int']['output']>;
    /** The ranking of the media in a particular time span and format compared to other media */
    rankings?: Maybe<Array<Maybe<MediaRank>>>;
    /** User recommendations for similar media */
    recommendations?: Maybe<RecommendationConnection>;
    /** Other media in the same or connecting franchise */
    relations?: Maybe<MediaConnection>;
    /** User reviews of the media */
    reviews?: Maybe<ReviewConnection>;
    /** The season the media was initially released in */
    season?: Maybe<MediaSeason>;
    /**
     * The year & season the media was initially released in
     * @deprecated
     */
    seasonInt?: Maybe<Scalars['Int']['output']>;
    /** The season year the media was initially released in */
    seasonYear?: Maybe<Scalars['Int']['output']>;
    /** The url for the media page on the AniList website */
    siteUrl?: Maybe<Scalars['String']['output']>;
    /** Source type the media was adapted from. */
    source?: Maybe<MediaSource>;
    /** The staff who produced the media */
    staff?: Maybe<StaffConnection>;
    /** The first official release date of the media */
    startDate?: Maybe<FuzzyDate>;
    stats?: Maybe<MediaStats>;
    /** The current releasing status of the media */
    status?: Maybe<MediaStatus>;
    /** Data and links to legal streaming episodes on external sites */
    streamingEpisodes?: Maybe<Array<Maybe<MediaStreamingEpisode>>>;
    /** The companies who produced the media */
    studios?: Maybe<StudioConnection>;
    /** Alternative titles of the media */
    synonyms?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
    /** List of tags that describes elements and themes of the media */
    tags?: Maybe<Array<Maybe<MediaTag>>>;
    /** The official titles of the media in various languages */
    title?: Maybe<MediaTitle>;
    /** Media trailer or advertisement */
    trailer?: Maybe<MediaTrailer>;
    /** The amount of related activity in the past hour */
    trending?: Maybe<Scalars['Int']['output']>;
    /** The media's daily trend stats */
    trends?: Maybe<MediaTrendConnection>;
    /** The type of the media; anime or manga */
    type?: Maybe<MediaType>;
    /** When the media's data was last updated */
    updatedAt?: Maybe<Scalars['Int']['output']>;
    /** The amount of volumes the manga has when complete */
    volumes?: Maybe<Scalars['Int']['output']>;
};


/** Anime or Manga */
export type MediaAiringScheduleArgs = {
    notYetAired?: InputMaybe<Scalars['Boolean']['input']>;
    page?: InputMaybe<Scalars['Int']['input']>;
    perPage?: InputMaybe<Scalars['Int']['input']>;
};


/** Anime or Manga */
export type MediaCharactersArgs = {
    page?: InputMaybe<Scalars['Int']['input']>;
    perPage?: InputMaybe<Scalars['Int']['input']>;
    role?: InputMaybe<CharacterRole>;
    sort?: InputMaybe<Array<InputMaybe<CharacterSort>>>;
};


/** Anime or Manga */
export type MediaDescriptionArgs = {
    asHtml?: InputMaybe<Scalars['Boolean']['input']>;
};


/** Anime or Manga */
export type MediaRecommendationsArgs = {
    page?: InputMaybe<Scalars['Int']['input']>;
    perPage?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<RecommendationSort>>>;
};


/** Anime or Manga */
export type MediaReviewsArgs = {
    limit?: InputMaybe<Scalars['Int']['input']>;
    page?: InputMaybe<Scalars['Int']['input']>;
    perPage?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<ReviewSort>>>;
};


/** Anime or Manga */
export type MediaSourceArgs = {
    version?: InputMaybe<Scalars['Int']['input']>;
};


/** Anime or Manga */
export type MediaStaffArgs = {
    page?: InputMaybe<Scalars['Int']['input']>;
    perPage?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<StaffSort>>>;
};


/** Anime or Manga */
export type MediaStatusArgs = {
    version?: InputMaybe<Scalars['Int']['input']>;
};


/** Anime or Manga */
export type MediaStudiosArgs = {
    isMain?: InputMaybe<Scalars['Boolean']['input']>;
    sort?: InputMaybe<Array<InputMaybe<StudioSort>>>;
};


/** Anime or Manga */
export type MediaTrendsArgs = {
    page?: InputMaybe<Scalars['Int']['input']>;
    perPage?: InputMaybe<Scalars['Int']['input']>;
    releasing?: InputMaybe<Scalars['Boolean']['input']>;
    sort?: InputMaybe<Array<InputMaybe<MediaTrendSort>>>;
};

/** Internal - Media characters separated */
export type MediaCharacter = {
    /** The characters in the media voiced by the parent actor */
    character?: Maybe<Character>;
    /** Media specific character name */
    characterName?: Maybe<Scalars['String']['output']>;
    dubGroup?: Maybe<Scalars['String']['output']>;
    /** The id of the connection */
    id?: Maybe<Scalars['Int']['output']>;
    /** The characters role in the media */
    role?: Maybe<CharacterRole>;
    roleNotes?: Maybe<Scalars['String']['output']>;
    /** The voice actor of the character */
    voiceActor?: Maybe<Staff>;
};

export type MediaConnection = {
    edges?: Maybe<Array<Maybe<MediaEdge>>>;
    nodes?: Maybe<Array<Maybe<Media>>>;
    /** The pagination information */
    pageInfo?: Maybe<PageInfo>;
};

export type MediaCoverImage = {
    /** Average #hex color of cover image */
    color?: Maybe<Scalars['String']['output']>;
    /** The cover image url of the media at its largest size. If this size isn't available, large will be provided instead. */
    extraLarge?: Maybe<Scalars['String']['output']>;
    /** The cover image url of the media at a large size */
    large?: Maybe<Scalars['String']['output']>;
    /** The cover image url of the media at medium size */
    medium?: Maybe<Scalars['String']['output']>;
};

/** Notification for when a media entry's data was changed in a significant way impacting users' list tracking */
export type MediaDataChangeNotification = {
    /** The reason for the media data change */
    context?: Maybe<Scalars['String']['output']>;
    /** The time the notification was created at */
    createdAt?: Maybe<Scalars['Int']['output']>;
    /** The id of the Notification */
    id: Scalars['Int']['output'];
    /** The media that received data changes */
    media?: Maybe<Media>;
    /** The id of the media that received data changes */
    mediaId: Scalars['Int']['output'];
    /** The reason for the media data change */
    reason?: Maybe<Scalars['String']['output']>;
    /** The type of notification */
    type?: Maybe<NotificationType>;
};

/** Notification for when a media tracked in a user's list is deleted from the site */
export type MediaDeletionNotification = {
    /** The reason for the media deletion */
    context?: Maybe<Scalars['String']['output']>;
    /** The time the notification was created at */
    createdAt?: Maybe<Scalars['Int']['output']>;
    /** The title of the deleted media */
    deletedMediaTitle?: Maybe<Scalars['String']['output']>;
    /** The id of the Notification */
    id: Scalars['Int']['output'];
    /** The reason for the media deletion */
    reason?: Maybe<Scalars['String']['output']>;
    /** The type of notification */
    type?: Maybe<NotificationType>;
};

/** Media connection edge */
export type MediaEdge = {
    /** Media specific character name */
    characterName?: Maybe<Scalars['String']['output']>;
    /** The characters role in the media */
    characterRole?: Maybe<CharacterRole>;
    /** The characters in the media voiced by the parent actor */
    characters?: Maybe<Array<Maybe<Character>>>;
    /** Used for grouping roles where multiple dubs exist for the same language. Either dubbing company name or language variant. */
    dubGroup?: Maybe<Scalars['String']['output']>;
    /** The order the media should be displayed from the users favourites */
    favouriteOrder?: Maybe<Scalars['Int']['output']>;
    /** The id of the connection */
    id?: Maybe<Scalars['Int']['output']>;
    /** If the studio is the main animation studio of the media (For Studio->MediaConnection field only) */
    isMainStudio: Scalars['Boolean']['output'];
    node?: Maybe<Media>;
    /** The type of relation to the parent model */
    relationType?: Maybe<MediaRelation>;
    /** Notes regarding the VA's role for the character */
    roleNotes?: Maybe<Scalars['String']['output']>;
    /** The role of the staff member in the production of the media */
    staffRole?: Maybe<Scalars['String']['output']>;
    /** The voice actors of the character with role date */
    voiceActorRoles?: Maybe<Array<Maybe<StaffRoleType>>>;
    /** The voice actors of the character */
    voiceActors?: Maybe<Array<Maybe<Staff>>>;
};


/** Media connection edge */
export type MediaEdgeRelationTypeArgs = {
    version?: InputMaybe<Scalars['Int']['input']>;
};


/** Media connection edge */
export type MediaEdgeVoiceActorRolesArgs = {
    language?: InputMaybe<StaffLanguage>;
    sort?: InputMaybe<Array<InputMaybe<StaffSort>>>;
};


/** Media connection edge */
export type MediaEdgeVoiceActorsArgs = {
    language?: InputMaybe<StaffLanguage>;
    sort?: InputMaybe<Array<InputMaybe<StaffSort>>>;
};

/** An external link to another site related to the media or staff member */
export type MediaExternalLink = {
    color?: Maybe<Scalars['String']['output']>;
    /** The icon image url of the site. Not available for all links. Transparent PNG 64x64 */
    icon?: Maybe<Scalars['String']['output']>;
    /** The id of the external link */
    id: Scalars['Int']['output'];
    isDisabled?: Maybe<Scalars['Boolean']['output']>;
    /** Language the site content is in. See Staff language field for values. */
    language?: Maybe<Scalars['String']['output']>;
    notes?: Maybe<Scalars['String']['output']>;
    /** The links website site name */
    site: Scalars['String']['output'];
    /** The links website site id */
    siteId?: Maybe<Scalars['Int']['output']>;
    type?: Maybe<ExternalLinkType>;
    /** The url of the external link or base url of link source */
    url?: Maybe<Scalars['String']['output']>;
};

/** An external link to another site related to the media */
export type MediaExternalLinkInput = {
    /** The id of the external link */
    id: Scalars['Int']['input'];
    /** The site location of the external link */
    site: Scalars['String']['input'];
    /** The url of the external link */
    url: Scalars['String']['input'];
};

/** The format the media was released in */
export type MediaFormat =
/** Professionally published manga with more than one chapter */
    | 'MANGA'
    /** Anime movies with a theatrical release */
    | 'MOVIE'
    /** Short anime released as a music video */
    | 'MUSIC'
    /** Written books released as a series of light novels */
    | 'NOVEL'
    /** (Original Net Animation) Anime that have been originally released online or are only available through streaming services. */
    | 'ONA'
    /** Manga with just one chapter */
    | 'ONE_SHOT'
    /** (Original Video Animation) Anime that have been released directly on DVD/Blu-ray without originally going through a theatrical release or television broadcast */
    | 'OVA'
    /** Special episodes that have been included in DVD/Blu-ray releases, picture dramas, pilots, etc */
    | 'SPECIAL'
    /** Anime broadcast on television */
    | 'TV'
    /** Anime which are under 15 minutes in length and broadcast on television */
    | 'TV_SHORT';

/** List of anime or manga */
export type MediaList = {
    /** Map of advanced scores with name keys */
    advancedScores?: Maybe<Scalars['Json']['output']>;
    /** When the entry was completed by the user */
    completedAt?: Maybe<FuzzyDate>;
    /** When the entry data was created */
    createdAt?: Maybe<Scalars['Int']['output']>;
    /** Map of booleans for which custom lists the entry are in */
    customLists?: Maybe<Scalars['Json']['output']>;
    /** If the entry shown be hidden from non-custom lists */
    hiddenFromStatusLists?: Maybe<Scalars['Boolean']['output']>;
    /** The id of the list entry */
    id: Scalars['Int']['output'];
    media?: Maybe<Media>;
    /** The id of the media */
    mediaId: Scalars['Int']['output'];
    /** Text notes */
    notes?: Maybe<Scalars['String']['output']>;
    /** Priority of planning */
    priority?: Maybe<Scalars['Int']['output']>;
    /** If the entry should only be visible to authenticated user */
    private?: Maybe<Scalars['Boolean']['output']>;
    /** The amount of episodes/chapters consumed by the user */
    progress?: Maybe<Scalars['Int']['output']>;
    /** The amount of volumes read by the user */
    progressVolumes?: Maybe<Scalars['Int']['output']>;
    /** The amount of times the user has rewatched/read the media */
    repeat?: Maybe<Scalars['Int']['output']>;
    /** The score of the entry */
    score?: Maybe<Scalars['Float']['output']>;
    /** When the entry was started by the user */
    startedAt?: Maybe<FuzzyDate>;
    /** The watching/reading status */
    status?: Maybe<MediaListStatus>;
    /** When the entry data was last updated */
    updatedAt?: Maybe<Scalars['Int']['output']>;
    user?: Maybe<User>;
    /** The id of the user owner of the list entry */
    userId: Scalars['Int']['output'];
};


/** List of anime or manga */
export type MediaListCustomListsArgs = {
    asArray?: InputMaybe<Scalars['Boolean']['input']>;
};


/** List of anime or manga */
export type MediaListScoreArgs = {
    format?: InputMaybe<ScoreFormat>;
};

/** List of anime or manga */
export type MediaListCollection = {
    /**
     * A map of media list entry arrays grouped by custom lists
     * @deprecated Not GraphQL spec compliant, use lists field instead.
     */
    customLists?: Maybe<Array<Maybe<Array<Maybe<MediaList>>>>>;
    /** If there is another chunk */
    hasNextChunk?: Maybe<Scalars['Boolean']['output']>;
    /** Grouped media list entries */
    lists?: Maybe<Array<Maybe<MediaListGroup>>>;
    /**
     * A map of media list entry arrays grouped by status
     * @deprecated Not GraphQL spec compliant, use lists field instead.
     */
    statusLists?: Maybe<Array<Maybe<Array<Maybe<MediaList>>>>>;
    /** The owner of the list */
    user?: Maybe<User>;
};


/** List of anime or manga */
export type MediaListCollectionCustomListsArgs = {
    asArray?: InputMaybe<Scalars['Boolean']['input']>;
};


/** List of anime or manga */
export type MediaListCollectionStatusListsArgs = {
    asArray?: InputMaybe<Scalars['Boolean']['input']>;
};

/** List group of anime or manga entries */
export type MediaListGroup = {
    /** Media list entries */
    entries?: Maybe<Array<Maybe<MediaList>>>;
    isCustomList?: Maybe<Scalars['Boolean']['output']>;
    isSplitCompletedList?: Maybe<Scalars['Boolean']['output']>;
    name?: Maybe<Scalars['String']['output']>;
    status?: Maybe<MediaListStatus>;
};

/** A user's list options */
export type MediaListOptions = {
    /** The user's anime list options */
    animeList?: Maybe<MediaListTypeOptions>;
    /** The user's manga list options */
    mangaList?: Maybe<MediaListTypeOptions>;
    /** The default order list rows should be displayed in */
    rowOrder?: Maybe<Scalars['String']['output']>;
    /** The score format the user is using for media lists */
    scoreFormat?: Maybe<ScoreFormat>;
    /**
     * The list theme options for both lists
     * @deprecated No longer used
     */
    sharedTheme?: Maybe<Scalars['Json']['output']>;
    /**
     * If the shared theme should be used instead of the individual list themes
     * @deprecated No longer used
     */
    sharedThemeEnabled?: Maybe<Scalars['Boolean']['output']>;
    /** @deprecated No longer used */
    useLegacyLists?: Maybe<Scalars['Boolean']['output']>;
};

/** A user's list options for anime or manga lists */
export type MediaListOptionsInput = {
    /** The names of the user's advanced scoring sections */
    advancedScoring?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    /** If advanced scoring is enabled */
    advancedScoringEnabled?: InputMaybe<Scalars['Boolean']['input']>;
    /** The names of the user's custom lists */
    customLists?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    /** The order each list should be displayed in */
    sectionOrder?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    /** If the completed sections of the list should be separated by format */
    splitCompletedSectionByFormat?: InputMaybe<Scalars['Boolean']['input']>;
    /** list theme */
    theme?: InputMaybe<Scalars['String']['input']>;
};

/** Media list sort enums */
export type MediaListSort =
    | 'ADDED_TIME'
    | 'ADDED_TIME_DESC'
    | 'FINISHED_ON'
    | 'FINISHED_ON_DESC'
    | 'MEDIA_ID'
    | 'MEDIA_ID_DESC'
    | 'MEDIA_POPULARITY'
    | 'MEDIA_POPULARITY_DESC'
    | 'MEDIA_TITLE_ENGLISH'
    | 'MEDIA_TITLE_ENGLISH_DESC'
    | 'MEDIA_TITLE_NATIVE'
    | 'MEDIA_TITLE_NATIVE_DESC'
    | 'MEDIA_TITLE_ROMAJI'
    | 'MEDIA_TITLE_ROMAJI_DESC'
    | 'PRIORITY'
    | 'PRIORITY_DESC'
    | 'PROGRESS'
    | 'PROGRESS_DESC'
    | 'PROGRESS_VOLUMES'
    | 'PROGRESS_VOLUMES_DESC'
    | 'REPEAT'
    | 'REPEAT_DESC'
    | 'SCORE'
    | 'SCORE_DESC'
    | 'STARTED_ON'
    | 'STARTED_ON_DESC'
    | 'STATUS'
    | 'STATUS_DESC'
    | 'UPDATED_TIME'
    | 'UPDATED_TIME_DESC';

/** Media list watching/reading status enum. */
export type MediaListStatus =
/** Finished watching/reading */
    | 'COMPLETED'
    /** Currently watching/reading */
    | 'CURRENT'
    /** Stopped watching/reading before completing */
    | 'DROPPED'
    /** Paused watching/reading */
    | 'PAUSED'
    /** Planning to watch/read */
    | 'PLANNING'
    /** Re-watching/reading */
    | 'REPEATING';

/** A user's list options for anime or manga lists */
export type MediaListTypeOptions = {
    /** The names of the user's advanced scoring sections */
    advancedScoring?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
    /** If advanced scoring is enabled */
    advancedScoringEnabled?: Maybe<Scalars['Boolean']['output']>;
    /** The names of the user's custom lists */
    customLists?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
    /** The order each list should be displayed in */
    sectionOrder?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
    /** If the completed sections of the list should be separated by format */
    splitCompletedSectionByFormat?: Maybe<Scalars['Boolean']['output']>;
    /**
     * The list theme options
     * @deprecated This field has not yet been fully implemented and may change without warning
     */
    theme?: Maybe<Scalars['Json']['output']>;
};

/** Notification for when a media entry is merged into another for a user who had it on their list */
export type MediaMergeNotification = {
    /** The reason for the media data change */
    context?: Maybe<Scalars['String']['output']>;
    /** The time the notification was created at */
    createdAt?: Maybe<Scalars['Int']['output']>;
    /** The title of the deleted media */
    deletedMediaTitles?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
    /** The id of the Notification */
    id: Scalars['Int']['output'];
    /** The media that was merged into */
    media?: Maybe<Media>;
    /** The id of the media that was merged into */
    mediaId: Scalars['Int']['output'];
    /** The reason for the media merge */
    reason?: Maybe<Scalars['String']['output']>;
    /** The type of notification */
    type?: Maybe<NotificationType>;
};

/** The ranking of a media in a particular time span and format compared to other media */
export type MediaRank = {
    /** If the ranking is based on all time instead of a season/year */
    allTime?: Maybe<Scalars['Boolean']['output']>;
    /** String that gives context to the ranking type and time span */
    context: Scalars['String']['output'];
    /** The format the media is ranked within */
    format: MediaFormat;
    /** The id of the rank */
    id: Scalars['Int']['output'];
    /** The numerical rank of the media */
    rank: Scalars['Int']['output'];
    /** The season the media is ranked within */
    season?: Maybe<MediaSeason>;
    /** The type of ranking */
    type: MediaRankType;
    /** The year the media is ranked within */
    year?: Maybe<Scalars['Int']['output']>;
};

/** The type of ranking */
export type MediaRankType =
/** Ranking is based on the media's popularity */
    | 'POPULAR'
    /** Ranking is based on the media's ratings/score */
    | 'RATED';

/** Type of relation media has to its parent. */
export type MediaRelation =
/** An adaption of this media into a different format */
    | 'ADAPTATION'
    /** An alternative version of the same media */
    | 'ALTERNATIVE'
    /** Shares at least 1 character */
    | 'CHARACTER'
    /** Version 2 only. */
    | 'COMPILATION'
    /** Version 2 only. */
    | 'CONTAINS'
    /** Other */
    | 'OTHER'
    /** The media a side story is from */
    | 'PARENT'
    /** Released before the relation */
    | 'PREQUEL'
    /** Released after the relation */
    | 'SEQUEL'
    /** A side story of the parent media */
    | 'SIDE_STORY'
    /** Version 2 only. The source material the media was adapted from */
    | 'SOURCE'
    /** An alternative version of the media with a different primary focus */
    | 'SPIN_OFF'
    /** A shortened and summarized version */
    | 'SUMMARY';

export type MediaSeason =
/** Months September to November */
    | 'FALL'
    /** Months March to May */
    | 'SPRING'
    /** Months June to August */
    | 'SUMMER'
    /** Months December to February */
    | 'WINTER';

/** Media sort enums */
export type MediaSort =
    | 'CHAPTERS'
    | 'CHAPTERS_DESC'
    | 'DURATION'
    | 'DURATION_DESC'
    | 'END_DATE'
    | 'END_DATE_DESC'
    | 'EPISODES'
    | 'EPISODES_DESC'
    | 'FAVOURITES'
    | 'FAVOURITES_DESC'
    | 'FORMAT'
    | 'FORMAT_DESC'
    | 'ID'
    | 'ID_DESC'
    | 'POPULARITY'
    | 'POPULARITY_DESC'
    | 'SCORE'
    | 'SCORE_DESC'
    | 'SEARCH_MATCH'
    | 'START_DATE'
    | 'START_DATE_DESC'
    | 'STATUS'
    | 'STATUS_DESC'
    | 'TITLE_ENGLISH'
    | 'TITLE_ENGLISH_DESC'
    | 'TITLE_NATIVE'
    | 'TITLE_NATIVE_DESC'
    | 'TITLE_ROMAJI'
    | 'TITLE_ROMAJI_DESC'
    | 'TRENDING'
    | 'TRENDING_DESC'
    | 'TYPE'
    | 'TYPE_DESC'
    | 'UPDATED_AT'
    | 'UPDATED_AT_DESC'
    | 'VOLUMES'
    | 'VOLUMES_DESC';

/** Source type the media was adapted from */
export type MediaSource =
/** Version 2+ only. Japanese Anime */
    | 'ANIME'
    /** Version 3 only. Comics excluding manga */
    | 'COMIC'
    /** Version 2+ only. Self-published works */
    | 'DOUJINSHI'
    /** Version 3 only. Games excluding video games */
    | 'GAME'
    /** Written work published in volumes */
    | 'LIGHT_NOVEL'
    /** Version 3 only. Live action media such as movies or TV show */
    | 'LIVE_ACTION'
    /** Asian comic book */
    | 'MANGA'
    /** Version 3 only. Multimedia project */
    | 'MULTIMEDIA_PROJECT'
    /** Version 2+ only. Written works not published in volumes */
    | 'NOVEL'
    /** An original production not based of another work */
    | 'ORIGINAL'
    /** Other */
    | 'OTHER'
    /** Version 3 only. Picture book */
    | 'PICTURE_BOOK'
    /** Video game */
    | 'VIDEO_GAME'
    /** Video game driven primary by text and narrative */
    | 'VISUAL_NOVEL'
    /** Version 3 only. Written works published online */
    | 'WEB_NOVEL';

/** A media's statistics */
export type MediaStats = {
    /** @deprecated Replaced by MediaTrends */
    airingProgression?: Maybe<Array<Maybe<AiringProgression>>>;
    scoreDistribution?: Maybe<Array<Maybe<ScoreDistribution>>>;
    statusDistribution?: Maybe<Array<Maybe<StatusDistribution>>>;
};

/** The current releasing status of the media */
export type MediaStatus =
/** Ended before the work could be finished */
    | 'CANCELLED'
    /** Has completed and is no longer being released */
    | 'FINISHED'
    /** Version 2 only. Is currently paused from releasing and will resume at a later date */
    | 'HIATUS'
    /** To be released at a later date */
    | 'NOT_YET_RELEASED'
    /** Currently releasing */
    | 'RELEASING';

/** Data and links to legal streaming episodes on external sites */
export type MediaStreamingEpisode = {
    /** The site location of the streaming episodes */
    site?: Maybe<Scalars['String']['output']>;
    /** Url of episode image thumbnail */
    thumbnail?: Maybe<Scalars['String']['output']>;
    /** Title of the episode */
    title?: Maybe<Scalars['String']['output']>;
    /** The url of the episode */
    url?: Maybe<Scalars['String']['output']>;
};

/** Media submission */
export type MediaSubmission = {
    /** Data Mod assigned to handle the submission */
    assignee?: Maybe<User>;
    changes?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
    characters?: Maybe<Array<Maybe<MediaSubmissionComparison>>>;
    createdAt?: Maybe<Scalars['Int']['output']>;
    externalLinks?: Maybe<Array<Maybe<MediaSubmissionComparison>>>;
    /** The id of the submission */
    id: Scalars['Int']['output'];
    /** Whether the submission is locked */
    locked?: Maybe<Scalars['Boolean']['output']>;
    media?: Maybe<Media>;
    notes?: Maybe<Scalars['String']['output']>;
    relations?: Maybe<Array<Maybe<MediaEdge>>>;
    source?: Maybe<Scalars['String']['output']>;
    staff?: Maybe<Array<Maybe<MediaSubmissionComparison>>>;
    /** Status of the submission */
    status?: Maybe<SubmissionStatus>;
    studios?: Maybe<Array<Maybe<MediaSubmissionComparison>>>;
    submission?: Maybe<Media>;
    /** User submitter of the submission */
    submitter?: Maybe<User>;
    submitterStats?: Maybe<Scalars['Json']['output']>;
};

/** Media submission with comparison to current data */
export type MediaSubmissionComparison = {
    character?: Maybe<MediaCharacter>;
    externalLink?: Maybe<MediaExternalLink>;
    staff?: Maybe<StaffEdge>;
    studio?: Maybe<StudioEdge>;
    submission?: Maybe<MediaSubmissionEdge>;
};

export type MediaSubmissionEdge = {
    character?: Maybe<Character>;
    characterName?: Maybe<Scalars['String']['output']>;
    characterRole?: Maybe<CharacterRole>;
    characterSubmission?: Maybe<Character>;
    dubGroup?: Maybe<Scalars['String']['output']>;
    externalLink?: Maybe<MediaExternalLink>;
    /** The id of the direct submission */
    id?: Maybe<Scalars['Int']['output']>;
    isMain?: Maybe<Scalars['Boolean']['output']>;
    media?: Maybe<Media>;
    roleNotes?: Maybe<Scalars['String']['output']>;
    staff?: Maybe<Staff>;
    staffRole?: Maybe<Scalars['String']['output']>;
    staffSubmission?: Maybe<Staff>;
    studio?: Maybe<Studio>;
    voiceActor?: Maybe<Staff>;
    voiceActorSubmission?: Maybe<Staff>;
};

/** A tag that describes a theme or element of the media */
export type MediaTag = {
    /** The categories of tags this tag belongs to */
    category?: Maybe<Scalars['String']['output']>;
    /** A general description of the tag */
    description?: Maybe<Scalars['String']['output']>;
    /** The id of the tag */
    id: Scalars['Int']['output'];
    /** If the tag is only for adult 18+ media */
    isAdult?: Maybe<Scalars['Boolean']['output']>;
    /** If the tag could be a spoiler for any media */
    isGeneralSpoiler?: Maybe<Scalars['Boolean']['output']>;
    /** If the tag is a spoiler for this media */
    isMediaSpoiler?: Maybe<Scalars['Boolean']['output']>;
    /** The name of the tag */
    name: Scalars['String']['output'];
    /** The relevance ranking of the tag out of the 100 for this media */
    rank?: Maybe<Scalars['Int']['output']>;
    /** The user who submitted the tag */
    userId?: Maybe<Scalars['Int']['output']>;
};

/** The official titles of the media in various languages */
export type MediaTitle = {
    /** The official english title */
    english?: Maybe<Scalars['String']['output']>;
    /** Official title in it's native language */
    native?: Maybe<Scalars['String']['output']>;
    /** The romanization of the native language title */
    romaji?: Maybe<Scalars['String']['output']>;
    /** The currently authenticated users preferred title language. Default romaji for non-authenticated */
    userPreferred?: Maybe<Scalars['String']['output']>;
};


/** The official titles of the media in various languages */
export type MediaTitleEnglishArgs = {
    stylised?: InputMaybe<Scalars['Boolean']['input']>;
};


/** The official titles of the media in various languages */
export type MediaTitleNativeArgs = {
    stylised?: InputMaybe<Scalars['Boolean']['input']>;
};


/** The official titles of the media in various languages */
export type MediaTitleRomajiArgs = {
    stylised?: InputMaybe<Scalars['Boolean']['input']>;
};

/** The official titles of the media in various languages */
export type MediaTitleInput = {
    /** The official english title */
    english?: InputMaybe<Scalars['String']['input']>;
    /** Official title in it's native language */
    native?: InputMaybe<Scalars['String']['input']>;
    /** The romanization of the native language title */
    romaji?: InputMaybe<Scalars['String']['input']>;
};

/** Media trailer or advertisement */
export type MediaTrailer = {
    /** The trailer video id */
    id?: Maybe<Scalars['String']['output']>;
    /** The site the video is hosted by (Currently either youtube or dailymotion) */
    site?: Maybe<Scalars['String']['output']>;
    /** The url for the thumbnail image of the video */
    thumbnail?: Maybe<Scalars['String']['output']>;
};

/** Daily media statistics */
export type MediaTrend = {
    /** A weighted average score of all the user's scores of the media */
    averageScore?: Maybe<Scalars['Int']['output']>;
    /** The day the data was recorded (timestamp) */
    date: Scalars['Int']['output'];
    /** The episode number of the anime released on this day */
    episode?: Maybe<Scalars['Int']['output']>;
    /** The number of users with watching/reading the media */
    inProgress?: Maybe<Scalars['Int']['output']>;
    /** The related media */
    media?: Maybe<Media>;
    /** The id of the tag */
    mediaId: Scalars['Int']['output'];
    /** The number of users with the media on their list */
    popularity?: Maybe<Scalars['Int']['output']>;
    /** If the media was being released at this time */
    releasing: Scalars['Boolean']['output'];
    /** The amount of media activity on the day */
    trending: Scalars['Int']['output'];
};

export type MediaTrendConnection = {
    edges?: Maybe<Array<Maybe<MediaTrendEdge>>>;
    nodes?: Maybe<Array<Maybe<MediaTrend>>>;
    /** The pagination information */
    pageInfo?: Maybe<PageInfo>;
};

/** Media trend connection edge */
export type MediaTrendEdge = {
    node?: Maybe<MediaTrend>;
};

/** Media trend sort enums */
export type MediaTrendSort =
    | 'DATE'
    | 'DATE_DESC'
    | 'EPISODE'
    | 'EPISODE_DESC'
    | 'ID'
    | 'ID_DESC'
    | 'MEDIA_ID'
    | 'MEDIA_ID_DESC'
    | 'POPULARITY'
    | 'POPULARITY_DESC'
    | 'SCORE'
    | 'SCORE_DESC'
    | 'TRENDING'
    | 'TRENDING_DESC';

/** Media type enum, anime or manga. */
export type MediaType =
/** Japanese Anime */
    | 'ANIME'
    /** Asian comic */
    | 'MANGA';

/** User message activity */
export type MessageActivity = {
    /** The time the activity was created at */
    createdAt: Scalars['Int']['output'];
    /** The id of the activity */
    id: Scalars['Int']['output'];
    /** If the currently authenticated user liked the activity */
    isLiked?: Maybe<Scalars['Boolean']['output']>;
    /** If the activity is locked and can receive replies */
    isLocked?: Maybe<Scalars['Boolean']['output']>;
    /** If the message is private and only viewable to the sender and recipients */
    isPrivate?: Maybe<Scalars['Boolean']['output']>;
    /** If the currently authenticated user is subscribed to the activity */
    isSubscribed?: Maybe<Scalars['Boolean']['output']>;
    /** The amount of likes the activity has */
    likeCount: Scalars['Int']['output'];
    /** The users who liked the activity */
    likes?: Maybe<Array<Maybe<User>>>;
    /** The message text (Markdown) */
    message?: Maybe<Scalars['String']['output']>;
    /** The user who sent the activity message */
    messenger?: Maybe<User>;
    /** The user id of the activity's sender */
    messengerId?: Maybe<Scalars['Int']['output']>;
    /** The user who the activity message was sent to */
    recipient?: Maybe<User>;
    /** The user id of the activity's recipient */
    recipientId?: Maybe<Scalars['Int']['output']>;
    /** The written replies to the activity */
    replies?: Maybe<Array<Maybe<ActivityReply>>>;
    /** The number of activity replies */
    replyCount: Scalars['Int']['output'];
    /** The url for the activity page on the AniList website */
    siteUrl?: Maybe<Scalars['String']['output']>;
    /** The type of the activity */
    type?: Maybe<ActivityType>;
};


/** User message activity */
export type MessageActivityMessageArgs = {
    asHtml?: InputMaybe<Scalars['Boolean']['input']>;
};

export type ModAction = {
    createdAt: Scalars['Int']['output'];
    data?: Maybe<Scalars['String']['output']>;
    /** The id of the action */
    id: Scalars['Int']['output'];
    mod?: Maybe<User>;
    objectId?: Maybe<Scalars['Int']['output']>;
    objectType?: Maybe<Scalars['String']['output']>;
    type?: Maybe<ModActionType>;
    user?: Maybe<User>;
};

export type ModActionType =
    | 'ANON'
    | 'BAN'
    | 'DELETE'
    | 'EDIT'
    | 'EXPIRE'
    | 'NOTE'
    | 'REPORT'
    | 'RESET';

/** Mod role enums */
export type ModRole =
/** An AniList administrator */
    | 'ADMIN'
    /** An anime data moderator */
    | 'ANIME_DATA'
    /** A character data moderator */
    | 'CHARACTER_DATA'
    /** A community moderator */
    | 'COMMUNITY'
    /** An AniList developer */
    | 'DEVELOPER'
    /** A discord community moderator */
    | 'DISCORD_COMMUNITY'
    /** A lead anime data moderator */
    | 'LEAD_ANIME_DATA'
    /** A lead community moderator */
    | 'LEAD_COMMUNITY'
    /** A head developer of AniList */
    | 'LEAD_DEVELOPER'
    /** A lead manga data moderator */
    | 'LEAD_MANGA_DATA'
    /** A lead social media moderator */
    | 'LEAD_SOCIAL_MEDIA'
    /** A manga data moderator */
    | 'MANGA_DATA'
    /** A retired moderator */
    | 'RETIRED'
    /** A social media moderator */
    | 'SOCIAL_MEDIA'
    /** A staff data moderator */
    | 'STAFF_DATA';

export type Mutation = {
    /** Delete an activity item of the authenticated users */
    DeleteActivity?: Maybe<Deleted>;
    /** Delete an activity reply of the authenticated users */
    DeleteActivityReply?: Maybe<Deleted>;
    /** Delete a custom list and remove the list entries from it */
    DeleteCustomList?: Maybe<Deleted>;
    /** Delete a media list entry */
    DeleteMediaListEntry?: Maybe<Deleted>;
    /** Delete a review */
    DeleteReview?: Maybe<Deleted>;
    /** Delete a thread */
    DeleteThread?: Maybe<Deleted>;
    /** Delete a thread comment */
    DeleteThreadComment?: Maybe<Deleted>;
    /** Rate a review */
    RateReview?: Maybe<Review>;
    /** Create or update an activity reply */
    SaveActivityReply?: Maybe<ActivityReply>;
    /** Update list activity (Mod Only) */
    SaveListActivity?: Maybe<ListActivity>;
    /** Create or update a media list entry */
    SaveMediaListEntry?: Maybe<MediaList>;
    /** Create or update message activity for the currently authenticated user */
    SaveMessageActivity?: Maybe<MessageActivity>;
    /** Recommendation a media */
    SaveRecommendation?: Maybe<Recommendation>;
    /** Create or update a review */
    SaveReview?: Maybe<Review>;
    /** Create or update text activity for the currently authenticated user */
    SaveTextActivity?: Maybe<TextActivity>;
    /** Create or update a forum thread */
    SaveThread?: Maybe<Thread>;
    /** Create or update a thread comment */
    SaveThreadComment?: Maybe<ThreadComment>;
    /** Toggle activity to be pinned to the top of the user's activity feed */
    ToggleActivityPin?: Maybe<ActivityUnion>;
    /** Toggle the subscription of an activity item */
    ToggleActivitySubscription?: Maybe<ActivityUnion>;
    /** Favourite or unfavourite an anime, manga, character, staff member, or studio */
    ToggleFavourite?: Maybe<Favourites>;
    /** Toggle the un/following of a user */
    ToggleFollow?: Maybe<User>;
    /**
     * Add or remove a like from a likeable type.
     *                           Returns all the users who liked the same model
     */
    ToggleLike?: Maybe<Array<Maybe<User>>>;
    /** Add or remove a like from a likeable type. */
    ToggleLikeV2?: Maybe<LikeableUnion>;
    /** Toggle the subscription of a forum thread */
    ToggleThreadSubscription?: Maybe<Thread>;
    UpdateAniChartHighlights?: Maybe<Scalars['Json']['output']>;
    UpdateAniChartSettings?: Maybe<Scalars['Json']['output']>;
    /** Update the order favourites are displayed in */
    UpdateFavouriteOrder?: Maybe<Favourites>;
    /** Update multiple media list entries to the same values */
    UpdateMediaListEntries?: Maybe<Array<Maybe<MediaList>>>;
    UpdateUser?: Maybe<User>;
};


export type MutationDeleteActivityArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
};


export type MutationDeleteActivityReplyArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
};


export type MutationDeleteCustomListArgs = {
    customList?: InputMaybe<Scalars['String']['input']>;
    type?: InputMaybe<MediaType>;
};


export type MutationDeleteMediaListEntryArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
};


export type MutationDeleteReviewArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
};


export type MutationDeleteThreadArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
};


export type MutationDeleteThreadCommentArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
};


export type MutationRateReviewArgs = {
    rating?: InputMaybe<ReviewRating>;
    reviewId?: InputMaybe<Scalars['Int']['input']>;
};


export type MutationSaveActivityReplyArgs = {
    activityId?: InputMaybe<Scalars['Int']['input']>;
    asMod?: InputMaybe<Scalars['Boolean']['input']>;
    id?: InputMaybe<Scalars['Int']['input']>;
    text?: InputMaybe<Scalars['String']['input']>;
};


export type MutationSaveListActivityArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
    locked?: InputMaybe<Scalars['Boolean']['input']>;
};


export type MutationSaveMediaListEntryArgs = {
    advancedScores?: InputMaybe<Array<InputMaybe<Scalars['Float']['input']>>>;
    completedAt?: InputMaybe<FuzzyDateInput>;
    customLists?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    hiddenFromStatusLists?: InputMaybe<Scalars['Boolean']['input']>;
    id?: InputMaybe<Scalars['Int']['input']>;
    mediaId?: InputMaybe<Scalars['Int']['input']>;
    notes?: InputMaybe<Scalars['String']['input']>;
    priority?: InputMaybe<Scalars['Int']['input']>;
    private?: InputMaybe<Scalars['Boolean']['input']>;
    progress?: InputMaybe<Scalars['Int']['input']>;
    progressVolumes?: InputMaybe<Scalars['Int']['input']>;
    repeat?: InputMaybe<Scalars['Int']['input']>;
    score?: InputMaybe<Scalars['Float']['input']>;
    scoreRaw?: InputMaybe<Scalars['Int']['input']>;
    startedAt?: InputMaybe<FuzzyDateInput>;
    status?: InputMaybe<MediaListStatus>;
};


export type MutationSaveMessageActivityArgs = {
    asMod?: InputMaybe<Scalars['Boolean']['input']>;
    id?: InputMaybe<Scalars['Int']['input']>;
    locked?: InputMaybe<Scalars['Boolean']['input']>;
    message?: InputMaybe<Scalars['String']['input']>;
    private?: InputMaybe<Scalars['Boolean']['input']>;
    recipientId?: InputMaybe<Scalars['Int']['input']>;
};


export type MutationSaveRecommendationArgs = {
    mediaId?: InputMaybe<Scalars['Int']['input']>;
    mediaRecommendationId?: InputMaybe<Scalars['Int']['input']>;
    rating?: InputMaybe<RecommendationRating>;
};


export type MutationSaveReviewArgs = {
    body?: InputMaybe<Scalars['String']['input']>;
    id?: InputMaybe<Scalars['Int']['input']>;
    mediaId?: InputMaybe<Scalars['Int']['input']>;
    private?: InputMaybe<Scalars['Boolean']['input']>;
    score?: InputMaybe<Scalars['Int']['input']>;
    summary?: InputMaybe<Scalars['String']['input']>;
};


export type MutationSaveTextActivityArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
    locked?: InputMaybe<Scalars['Boolean']['input']>;
    text?: InputMaybe<Scalars['String']['input']>;
};


export type MutationSaveThreadArgs = {
    body?: InputMaybe<Scalars['String']['input']>;
    categories?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    id?: InputMaybe<Scalars['Int']['input']>;
    locked?: InputMaybe<Scalars['Boolean']['input']>;
    mediaCategories?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    sticky?: InputMaybe<Scalars['Boolean']['input']>;
    title?: InputMaybe<Scalars['String']['input']>;
};


export type MutationSaveThreadCommentArgs = {
    comment?: InputMaybe<Scalars['String']['input']>;
    id?: InputMaybe<Scalars['Int']['input']>;
    locked?: InputMaybe<Scalars['Boolean']['input']>;
    parentCommentId?: InputMaybe<Scalars['Int']['input']>;
    threadId?: InputMaybe<Scalars['Int']['input']>;
};


export type MutationToggleActivityPinArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
    pinned?: InputMaybe<Scalars['Boolean']['input']>;
};


export type MutationToggleActivitySubscriptionArgs = {
    activityId?: InputMaybe<Scalars['Int']['input']>;
    subscribe?: InputMaybe<Scalars['Boolean']['input']>;
};


export type MutationToggleFavouriteArgs = {
    animeId?: InputMaybe<Scalars['Int']['input']>;
    characterId?: InputMaybe<Scalars['Int']['input']>;
    mangaId?: InputMaybe<Scalars['Int']['input']>;
    staffId?: InputMaybe<Scalars['Int']['input']>;
    studioId?: InputMaybe<Scalars['Int']['input']>;
};


export type MutationToggleFollowArgs = {
    userId?: InputMaybe<Scalars['Int']['input']>;
};


export type MutationToggleLikeArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
    type?: InputMaybe<LikeableType>;
};


export type MutationToggleLikeV2Args = {
    id?: InputMaybe<Scalars['Int']['input']>;
    type?: InputMaybe<LikeableType>;
};


export type MutationToggleThreadSubscriptionArgs = {
    subscribe?: InputMaybe<Scalars['Boolean']['input']>;
    threadId?: InputMaybe<Scalars['Int']['input']>;
};


export type MutationUpdateAniChartHighlightsArgs = {
    highlights?: InputMaybe<Array<InputMaybe<AniChartHighlightInput>>>;
};


export type MutationUpdateAniChartSettingsArgs = {
    outgoingLinkProvider?: InputMaybe<Scalars['String']['input']>;
    sort?: InputMaybe<Scalars['String']['input']>;
    theme?: InputMaybe<Scalars['String']['input']>;
    titleLanguage?: InputMaybe<Scalars['String']['input']>;
};


export type MutationUpdateFavouriteOrderArgs = {
    animeIds?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    animeOrder?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    characterIds?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    characterOrder?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    mangaIds?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    mangaOrder?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    staffIds?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    staffOrder?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    studioIds?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    studioOrder?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
};


export type MutationUpdateMediaListEntriesArgs = {
    advancedScores?: InputMaybe<Array<InputMaybe<Scalars['Float']['input']>>>;
    completedAt?: InputMaybe<FuzzyDateInput>;
    hiddenFromStatusLists?: InputMaybe<Scalars['Boolean']['input']>;
    ids?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    notes?: InputMaybe<Scalars['String']['input']>;
    priority?: InputMaybe<Scalars['Int']['input']>;
    private?: InputMaybe<Scalars['Boolean']['input']>;
    progress?: InputMaybe<Scalars['Int']['input']>;
    progressVolumes?: InputMaybe<Scalars['Int']['input']>;
    repeat?: InputMaybe<Scalars['Int']['input']>;
    score?: InputMaybe<Scalars['Float']['input']>;
    scoreRaw?: InputMaybe<Scalars['Int']['input']>;
    startedAt?: InputMaybe<FuzzyDateInput>;
    status?: InputMaybe<MediaListStatus>;
};


export type MutationUpdateUserArgs = {
    about?: InputMaybe<Scalars['String']['input']>;
    activityMergeTime?: InputMaybe<Scalars['Int']['input']>;
    airingNotifications?: InputMaybe<Scalars['Boolean']['input']>;
    animeListOptions?: InputMaybe<MediaListOptionsInput>;
    disabledListActivity?: InputMaybe<Array<InputMaybe<ListActivityOptionInput>>>;
    displayAdultContent?: InputMaybe<Scalars['Boolean']['input']>;
    donatorBadge?: InputMaybe<Scalars['String']['input']>;
    mangaListOptions?: InputMaybe<MediaListOptionsInput>;
    notificationOptions?: InputMaybe<Array<InputMaybe<NotificationOptionInput>>>;
    profileColor?: InputMaybe<Scalars['String']['input']>;
    restrictMessagesToFollowing?: InputMaybe<Scalars['Boolean']['input']>;
    rowOrder?: InputMaybe<Scalars['String']['input']>;
    scoreFormat?: InputMaybe<ScoreFormat>;
    staffNameLanguage?: InputMaybe<UserStaffNameLanguage>;
    timezone?: InputMaybe<Scalars['String']['input']>;
    titleLanguage?: InputMaybe<UserTitleLanguage>;
};

/** Notification option */
export type NotificationOption = {
    /** Whether this type of notification is enabled */
    enabled?: Maybe<Scalars['Boolean']['output']>;
    /** The type of notification */
    type?: Maybe<NotificationType>;
};

/** Notification option input */
export type NotificationOptionInput = {
    /** Whether this type of notification is enabled */
    enabled?: InputMaybe<Scalars['Boolean']['input']>;
    /** The type of notification */
    type?: InputMaybe<NotificationType>;
};

/** Notification type enum */
export type NotificationType =
/** A user has liked your activity */
    | 'ACTIVITY_LIKE'
    /** A user has mentioned you in their activity */
    | 'ACTIVITY_MENTION'
    /** A user has sent you message */
    | 'ACTIVITY_MESSAGE'
    /** A user has replied to your activity */
    | 'ACTIVITY_REPLY'
    /** A user has liked your activity reply */
    | 'ACTIVITY_REPLY_LIKE'
    /** A user has replied to activity you have also replied to */
    | 'ACTIVITY_REPLY_SUBSCRIBED'
    /** An anime you are currently watching has aired */
    | 'AIRING'
    /** A user has followed you */
    | 'FOLLOWING'
    /** An anime or manga has had a data change that affects how a user may track it in their lists */
    | 'MEDIA_DATA_CHANGE'
    /** An anime or manga on the user's list has been deleted from the site */
    | 'MEDIA_DELETION'
    /** Anime or manga entries on the user's list have been merged into a single entry */
    | 'MEDIA_MERGE'
    /** A new anime or manga has been added to the site where its related media is on the user's list */
    | 'RELATED_MEDIA_ADDITION'
    /** A user has liked your forum comment */
    | 'THREAD_COMMENT_LIKE'
    /** A user has mentioned you in a forum comment */
    | 'THREAD_COMMENT_MENTION'
    /** A user has replied to your forum comment */
    | 'THREAD_COMMENT_REPLY'
    /** A user has liked your forum thread */
    | 'THREAD_LIKE'
    /** A user has commented in one of your subscribed forum threads */
    | 'THREAD_SUBSCRIBED';

/** Notification union type */
export type NotificationUnion =
    ActivityLikeNotification
    | ActivityMentionNotification
    | ActivityMessageNotification
    | ActivityReplyLikeNotification
    | ActivityReplyNotification
    | ActivityReplySubscribedNotification
    | AiringNotification
    | FollowingNotification
    | MediaDataChangeNotification
    | MediaDeletionNotification
    | MediaMergeNotification
    | RelatedMediaAdditionNotification
    | ThreadCommentLikeNotification
    | ThreadCommentMentionNotification
    | ThreadCommentReplyNotification
    | ThreadCommentSubscribedNotification
    | ThreadLikeNotification;

/** Page of data */
export type Page = {
    activities?: Maybe<Array<Maybe<ActivityUnion>>>;
    activityReplies?: Maybe<Array<Maybe<ActivityReply>>>;
    airingSchedules?: Maybe<Array<Maybe<AiringSchedule>>>;
    characters?: Maybe<Array<Maybe<Character>>>;
    followers?: Maybe<Array<Maybe<User>>>;
    following?: Maybe<Array<Maybe<User>>>;
    likes?: Maybe<Array<Maybe<User>>>;
    media?: Maybe<Array<Maybe<Media>>>;
    mediaList?: Maybe<Array<Maybe<MediaList>>>;
    mediaTrends?: Maybe<Array<Maybe<MediaTrend>>>;
    notifications?: Maybe<Array<Maybe<NotificationUnion>>>;
    /** The pagination information */
    pageInfo?: Maybe<PageInfo>;
    recommendations?: Maybe<Array<Maybe<Recommendation>>>;
    reviews?: Maybe<Array<Maybe<Review>>>;
    staff?: Maybe<Array<Maybe<Staff>>>;
    studios?: Maybe<Array<Maybe<Studio>>>;
    threadComments?: Maybe<Array<Maybe<ThreadComment>>>;
    threads?: Maybe<Array<Maybe<Thread>>>;
    users?: Maybe<Array<Maybe<User>>>;
};


/** Page of data */
export type PageActivitiesArgs = {
    createdAt?: InputMaybe<Scalars['Int']['input']>;
    createdAt_greater?: InputMaybe<Scalars['Int']['input']>;
    createdAt_lesser?: InputMaybe<Scalars['Int']['input']>;
    hasReplies?: InputMaybe<Scalars['Boolean']['input']>;
    hasRepliesOrTypeText?: InputMaybe<Scalars['Boolean']['input']>;
    id?: InputMaybe<Scalars['Int']['input']>;
    id_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    id_not?: InputMaybe<Scalars['Int']['input']>;
    id_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    isFollowing?: InputMaybe<Scalars['Boolean']['input']>;
    mediaId?: InputMaybe<Scalars['Int']['input']>;
    mediaId_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    mediaId_not?: InputMaybe<Scalars['Int']['input']>;
    mediaId_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    messengerId?: InputMaybe<Scalars['Int']['input']>;
    messengerId_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    messengerId_not?: InputMaybe<Scalars['Int']['input']>;
    messengerId_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    sort?: InputMaybe<Array<InputMaybe<ActivitySort>>>;
    type?: InputMaybe<ActivityType>;
    type_in?: InputMaybe<Array<InputMaybe<ActivityType>>>;
    type_not?: InputMaybe<ActivityType>;
    type_not_in?: InputMaybe<Array<InputMaybe<ActivityType>>>;
    userId?: InputMaybe<Scalars['Int']['input']>;
    userId_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    userId_not?: InputMaybe<Scalars['Int']['input']>;
    userId_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
};


/** Page of data */
export type PageActivityRepliesArgs = {
    activityId?: InputMaybe<Scalars['Int']['input']>;
    id?: InputMaybe<Scalars['Int']['input']>;
};


/** Page of data */
export type PageAiringSchedulesArgs = {
    airingAt?: InputMaybe<Scalars['Int']['input']>;
    airingAt_greater?: InputMaybe<Scalars['Int']['input']>;
    airingAt_lesser?: InputMaybe<Scalars['Int']['input']>;
    episode?: InputMaybe<Scalars['Int']['input']>;
    episode_greater?: InputMaybe<Scalars['Int']['input']>;
    episode_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    episode_lesser?: InputMaybe<Scalars['Int']['input']>;
    episode_not?: InputMaybe<Scalars['Int']['input']>;
    episode_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    id?: InputMaybe<Scalars['Int']['input']>;
    id_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    id_not?: InputMaybe<Scalars['Int']['input']>;
    id_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    mediaId?: InputMaybe<Scalars['Int']['input']>;
    mediaId_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    mediaId_not?: InputMaybe<Scalars['Int']['input']>;
    mediaId_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    notYetAired?: InputMaybe<Scalars['Boolean']['input']>;
    sort?: InputMaybe<Array<InputMaybe<AiringSort>>>;
};


/** Page of data */
export type PageCharactersArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
    id_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    id_not?: InputMaybe<Scalars['Int']['input']>;
    id_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    isBirthday?: InputMaybe<Scalars['Boolean']['input']>;
    search?: InputMaybe<Scalars['String']['input']>;
    sort?: InputMaybe<Array<InputMaybe<CharacterSort>>>;
};


/** Page of data */
export type PageFollowersArgs = {
    sort?: InputMaybe<Array<InputMaybe<UserSort>>>;
    userId: Scalars['Int']['input'];
};


/** Page of data */
export type PageFollowingArgs = {
    sort?: InputMaybe<Array<InputMaybe<UserSort>>>;
    userId: Scalars['Int']['input'];
};


/** Page of data */
export type PageLikesArgs = {
    likeableId?: InputMaybe<Scalars['Int']['input']>;
    type?: InputMaybe<LikeableType>;
};


/** Page of data */
export type PageMediaArgs = {
    averageScore?: InputMaybe<Scalars['Int']['input']>;
    averageScore_greater?: InputMaybe<Scalars['Int']['input']>;
    averageScore_lesser?: InputMaybe<Scalars['Int']['input']>;
    averageScore_not?: InputMaybe<Scalars['Int']['input']>;
    chapters?: InputMaybe<Scalars['Int']['input']>;
    chapters_greater?: InputMaybe<Scalars['Int']['input']>;
    chapters_lesser?: InputMaybe<Scalars['Int']['input']>;
    countryOfOrigin?: InputMaybe<Scalars['CountryCode']['input']>;
    duration?: InputMaybe<Scalars['Int']['input']>;
    duration_greater?: InputMaybe<Scalars['Int']['input']>;
    duration_lesser?: InputMaybe<Scalars['Int']['input']>;
    endDate?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    endDate_greater?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    endDate_lesser?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    endDate_like?: InputMaybe<Scalars['String']['input']>;
    episodes?: InputMaybe<Scalars['Int']['input']>;
    episodes_greater?: InputMaybe<Scalars['Int']['input']>;
    episodes_lesser?: InputMaybe<Scalars['Int']['input']>;
    format?: InputMaybe<MediaFormat>;
    format_in?: InputMaybe<Array<InputMaybe<MediaFormat>>>;
    format_not?: InputMaybe<MediaFormat>;
    format_not_in?: InputMaybe<Array<InputMaybe<MediaFormat>>>;
    genre?: InputMaybe<Scalars['String']['input']>;
    genre_in?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    genre_not_in?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    id?: InputMaybe<Scalars['Int']['input']>;
    idMal?: InputMaybe<Scalars['Int']['input']>;
    idMal_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    idMal_not?: InputMaybe<Scalars['Int']['input']>;
    idMal_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    id_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    id_not?: InputMaybe<Scalars['Int']['input']>;
    id_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    isAdult?: InputMaybe<Scalars['Boolean']['input']>;
    isLicensed?: InputMaybe<Scalars['Boolean']['input']>;
    licensedBy?: InputMaybe<Scalars['String']['input']>;
    licensedById?: InputMaybe<Scalars['Int']['input']>;
    licensedById_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    licensedBy_in?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    minimumTagRank?: InputMaybe<Scalars['Int']['input']>;
    onList?: InputMaybe<Scalars['Boolean']['input']>;
    popularity?: InputMaybe<Scalars['Int']['input']>;
    popularity_greater?: InputMaybe<Scalars['Int']['input']>;
    popularity_lesser?: InputMaybe<Scalars['Int']['input']>;
    popularity_not?: InputMaybe<Scalars['Int']['input']>;
    search?: InputMaybe<Scalars['String']['input']>;
    season?: InputMaybe<MediaSeason>;
    seasonYear?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<MediaSort>>>;
    source?: InputMaybe<MediaSource>;
    source_in?: InputMaybe<Array<InputMaybe<MediaSource>>>;
    startDate?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    startDate_greater?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    startDate_lesser?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    startDate_like?: InputMaybe<Scalars['String']['input']>;
    status?: InputMaybe<MediaStatus>;
    status_in?: InputMaybe<Array<InputMaybe<MediaStatus>>>;
    status_not?: InputMaybe<MediaStatus>;
    status_not_in?: InputMaybe<Array<InputMaybe<MediaStatus>>>;
    tag?: InputMaybe<Scalars['String']['input']>;
    tagCategory?: InputMaybe<Scalars['String']['input']>;
    tagCategory_in?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    tagCategory_not_in?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    tag_in?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    tag_not_in?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    type?: InputMaybe<MediaType>;
    volumes?: InputMaybe<Scalars['Int']['input']>;
    volumes_greater?: InputMaybe<Scalars['Int']['input']>;
    volumes_lesser?: InputMaybe<Scalars['Int']['input']>;
};


/** Page of data */
export type PageMediaListArgs = {
    compareWithAuthList?: InputMaybe<Scalars['Boolean']['input']>;
    completedAt?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    completedAt_greater?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    completedAt_lesser?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    completedAt_like?: InputMaybe<Scalars['String']['input']>;
    id?: InputMaybe<Scalars['Int']['input']>;
    isFollowing?: InputMaybe<Scalars['Boolean']['input']>;
    mediaId?: InputMaybe<Scalars['Int']['input']>;
    mediaId_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    mediaId_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    notes?: InputMaybe<Scalars['String']['input']>;
    notes_like?: InputMaybe<Scalars['String']['input']>;
    sort?: InputMaybe<Array<InputMaybe<MediaListSort>>>;
    startedAt?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    startedAt_greater?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    startedAt_lesser?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    startedAt_like?: InputMaybe<Scalars['String']['input']>;
    status?: InputMaybe<MediaListStatus>;
    status_in?: InputMaybe<Array<InputMaybe<MediaListStatus>>>;
    status_not?: InputMaybe<MediaListStatus>;
    status_not_in?: InputMaybe<Array<InputMaybe<MediaListStatus>>>;
    type?: InputMaybe<MediaType>;
    userId?: InputMaybe<Scalars['Int']['input']>;
    userId_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    userName?: InputMaybe<Scalars['String']['input']>;
};


/** Page of data */
export type PageMediaTrendsArgs = {
    averageScore?: InputMaybe<Scalars['Int']['input']>;
    averageScore_greater?: InputMaybe<Scalars['Int']['input']>;
    averageScore_lesser?: InputMaybe<Scalars['Int']['input']>;
    averageScore_not?: InputMaybe<Scalars['Int']['input']>;
    date?: InputMaybe<Scalars['Int']['input']>;
    date_greater?: InputMaybe<Scalars['Int']['input']>;
    date_lesser?: InputMaybe<Scalars['Int']['input']>;
    episode?: InputMaybe<Scalars['Int']['input']>;
    episode_greater?: InputMaybe<Scalars['Int']['input']>;
    episode_lesser?: InputMaybe<Scalars['Int']['input']>;
    episode_not?: InputMaybe<Scalars['Int']['input']>;
    mediaId?: InputMaybe<Scalars['Int']['input']>;
    mediaId_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    mediaId_not?: InputMaybe<Scalars['Int']['input']>;
    mediaId_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    popularity?: InputMaybe<Scalars['Int']['input']>;
    popularity_greater?: InputMaybe<Scalars['Int']['input']>;
    popularity_lesser?: InputMaybe<Scalars['Int']['input']>;
    popularity_not?: InputMaybe<Scalars['Int']['input']>;
    releasing?: InputMaybe<Scalars['Boolean']['input']>;
    sort?: InputMaybe<Array<InputMaybe<MediaTrendSort>>>;
    trending?: InputMaybe<Scalars['Int']['input']>;
    trending_greater?: InputMaybe<Scalars['Int']['input']>;
    trending_lesser?: InputMaybe<Scalars['Int']['input']>;
    trending_not?: InputMaybe<Scalars['Int']['input']>;
};


/** Page of data */
export type PageNotificationsArgs = {
    resetNotificationCount?: InputMaybe<Scalars['Boolean']['input']>;
    type?: InputMaybe<NotificationType>;
    type_in?: InputMaybe<Array<InputMaybe<NotificationType>>>;
};


/** Page of data */
export type PageRecommendationsArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
    mediaId?: InputMaybe<Scalars['Int']['input']>;
    mediaRecommendationId?: InputMaybe<Scalars['Int']['input']>;
    onList?: InputMaybe<Scalars['Boolean']['input']>;
    rating?: InputMaybe<Scalars['Int']['input']>;
    rating_greater?: InputMaybe<Scalars['Int']['input']>;
    rating_lesser?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<RecommendationSort>>>;
    userId?: InputMaybe<Scalars['Int']['input']>;
};


/** Page of data */
export type PageReviewsArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
    mediaId?: InputMaybe<Scalars['Int']['input']>;
    mediaType?: InputMaybe<MediaType>;
    sort?: InputMaybe<Array<InputMaybe<ReviewSort>>>;
    userId?: InputMaybe<Scalars['Int']['input']>;
};


/** Page of data */
export type PageStaffArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
    id_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    id_not?: InputMaybe<Scalars['Int']['input']>;
    id_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    isBirthday?: InputMaybe<Scalars['Boolean']['input']>;
    search?: InputMaybe<Scalars['String']['input']>;
    sort?: InputMaybe<Array<InputMaybe<StaffSort>>>;
};


/** Page of data */
export type PageStudiosArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
    id_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    id_not?: InputMaybe<Scalars['Int']['input']>;
    id_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    search?: InputMaybe<Scalars['String']['input']>;
    sort?: InputMaybe<Array<InputMaybe<StudioSort>>>;
};


/** Page of data */
export type PageThreadCommentsArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<ThreadCommentSort>>>;
    threadId?: InputMaybe<Scalars['Int']['input']>;
    userId?: InputMaybe<Scalars['Int']['input']>;
};


/** Page of data */
export type PageThreadsArgs = {
    categoryId?: InputMaybe<Scalars['Int']['input']>;
    id?: InputMaybe<Scalars['Int']['input']>;
    id_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    mediaCategoryId?: InputMaybe<Scalars['Int']['input']>;
    replyUserId?: InputMaybe<Scalars['Int']['input']>;
    search?: InputMaybe<Scalars['String']['input']>;
    sort?: InputMaybe<Array<InputMaybe<ThreadSort>>>;
    subscribed?: InputMaybe<Scalars['Boolean']['input']>;
    userId?: InputMaybe<Scalars['Int']['input']>;
};


/** Page of data */
export type PageUsersArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
    isModerator?: InputMaybe<Scalars['Boolean']['input']>;
    name?: InputMaybe<Scalars['String']['input']>;
    search?: InputMaybe<Scalars['String']['input']>;
    sort?: InputMaybe<Array<InputMaybe<UserSort>>>;
};

export type PageInfo = {
    /** The current page */
    currentPage?: Maybe<Scalars['Int']['output']>;
    /** If there is another page */
    hasNextPage?: Maybe<Scalars['Boolean']['output']>;
    /** The last page */
    lastPage?: Maybe<Scalars['Int']['output']>;
    /** The count on a page */
    perPage?: Maybe<Scalars['Int']['output']>;
    /** The total number of items. Note: This value is not guaranteed to be accurate, do not rely on this for logic */
    total?: Maybe<Scalars['Int']['output']>;
};

/** Provides the parsed markdown as html */
export type ParsedMarkdown = {
    /** The parsed markdown as html */
    html?: Maybe<Scalars['String']['output']>;
};

export type Query = {
    /** Activity query */
    Activity?: Maybe<ActivityUnion>;
    /** Activity reply query */
    ActivityReply?: Maybe<ActivityReply>;
    /** Airing schedule query */
    AiringSchedule?: Maybe<AiringSchedule>;
    AniChartUser?: Maybe<AniChartUser>;
    /** Character query */
    Character?: Maybe<Character>;
    /** ExternalLinkSource collection query */
    ExternalLinkSourceCollection?: Maybe<Array<Maybe<MediaExternalLink>>>;
    /** Follow query */
    Follower?: Maybe<User>;
    /** Follow query */
    Following?: Maybe<User>;
    /** Collection of all the possible media genres */
    GenreCollection?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
    /** Like query */
    Like?: Maybe<User>;
    /** Provide AniList markdown to be converted to html (Requires auth) */
    Markdown?: Maybe<ParsedMarkdown>;
    /** Media query */
    Media?: Maybe<Media>;
    /** Media list query */
    MediaList?: Maybe<MediaList>;
    /** Media list collection query, provides list pre-grouped by status & custom lists. User ID and Media Type arguments required. */
    MediaListCollection?: Maybe<MediaListCollection>;
    /** Collection of all the possible media tags */
    MediaTagCollection?: Maybe<Array<Maybe<MediaTag>>>;
    /** Media Trend query */
    MediaTrend?: Maybe<MediaTrend>;
    /** Notification query */
    Notification?: Maybe<NotificationUnion>;
    Page?: Maybe<Page>;
    /** Recommendation query */
    Recommendation?: Maybe<Recommendation>;
    /** Review query */
    Review?: Maybe<Review>;
    /** Site statistics query */
    SiteStatistics?: Maybe<SiteStatistics>;
    /** Staff query */
    Staff?: Maybe<Staff>;
    /** Studio query */
    Studio?: Maybe<Studio>;
    /** Thread query */
    Thread?: Maybe<Thread>;
    /** Comment query */
    ThreadComment?: Maybe<Array<Maybe<ThreadComment>>>;
    /** User query */
    User?: Maybe<User>;
    /** Get the currently authenticated user */
    Viewer?: Maybe<User>;
};


export type QueryActivityArgs = {
    createdAt?: InputMaybe<Scalars['Int']['input']>;
    createdAt_greater?: InputMaybe<Scalars['Int']['input']>;
    createdAt_lesser?: InputMaybe<Scalars['Int']['input']>;
    hasReplies?: InputMaybe<Scalars['Boolean']['input']>;
    hasRepliesOrTypeText?: InputMaybe<Scalars['Boolean']['input']>;
    id?: InputMaybe<Scalars['Int']['input']>;
    id_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    id_not?: InputMaybe<Scalars['Int']['input']>;
    id_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    isFollowing?: InputMaybe<Scalars['Boolean']['input']>;
    mediaId?: InputMaybe<Scalars['Int']['input']>;
    mediaId_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    mediaId_not?: InputMaybe<Scalars['Int']['input']>;
    mediaId_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    messengerId?: InputMaybe<Scalars['Int']['input']>;
    messengerId_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    messengerId_not?: InputMaybe<Scalars['Int']['input']>;
    messengerId_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    sort?: InputMaybe<Array<InputMaybe<ActivitySort>>>;
    type?: InputMaybe<ActivityType>;
    type_in?: InputMaybe<Array<InputMaybe<ActivityType>>>;
    type_not?: InputMaybe<ActivityType>;
    type_not_in?: InputMaybe<Array<InputMaybe<ActivityType>>>;
    userId?: InputMaybe<Scalars['Int']['input']>;
    userId_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    userId_not?: InputMaybe<Scalars['Int']['input']>;
    userId_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
};


export type QueryActivityReplyArgs = {
    activityId?: InputMaybe<Scalars['Int']['input']>;
    id?: InputMaybe<Scalars['Int']['input']>;
};


export type QueryAiringScheduleArgs = {
    airingAt?: InputMaybe<Scalars['Int']['input']>;
    airingAt_greater?: InputMaybe<Scalars['Int']['input']>;
    airingAt_lesser?: InputMaybe<Scalars['Int']['input']>;
    episode?: InputMaybe<Scalars['Int']['input']>;
    episode_greater?: InputMaybe<Scalars['Int']['input']>;
    episode_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    episode_lesser?: InputMaybe<Scalars['Int']['input']>;
    episode_not?: InputMaybe<Scalars['Int']['input']>;
    episode_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    id?: InputMaybe<Scalars['Int']['input']>;
    id_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    id_not?: InputMaybe<Scalars['Int']['input']>;
    id_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    mediaId?: InputMaybe<Scalars['Int']['input']>;
    mediaId_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    mediaId_not?: InputMaybe<Scalars['Int']['input']>;
    mediaId_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    notYetAired?: InputMaybe<Scalars['Boolean']['input']>;
    sort?: InputMaybe<Array<InputMaybe<AiringSort>>>;
};


export type QueryCharacterArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
    id_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    id_not?: InputMaybe<Scalars['Int']['input']>;
    id_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    isBirthday?: InputMaybe<Scalars['Boolean']['input']>;
    search?: InputMaybe<Scalars['String']['input']>;
    sort?: InputMaybe<Array<InputMaybe<CharacterSort>>>;
};


export type QueryExternalLinkSourceCollectionArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
    mediaType?: InputMaybe<ExternalLinkMediaType>;
    type?: InputMaybe<ExternalLinkType>;
};


export type QueryFollowerArgs = {
    sort?: InputMaybe<Array<InputMaybe<UserSort>>>;
    userId: Scalars['Int']['input'];
};


export type QueryFollowingArgs = {
    sort?: InputMaybe<Array<InputMaybe<UserSort>>>;
    userId: Scalars['Int']['input'];
};


export type QueryLikeArgs = {
    likeableId?: InputMaybe<Scalars['Int']['input']>;
    type?: InputMaybe<LikeableType>;
};


export type QueryMarkdownArgs = {
    markdown: Scalars['String']['input'];
};


export type QueryMediaArgs = {
    averageScore?: InputMaybe<Scalars['Int']['input']>;
    averageScore_greater?: InputMaybe<Scalars['Int']['input']>;
    averageScore_lesser?: InputMaybe<Scalars['Int']['input']>;
    averageScore_not?: InputMaybe<Scalars['Int']['input']>;
    chapters?: InputMaybe<Scalars['Int']['input']>;
    chapters_greater?: InputMaybe<Scalars['Int']['input']>;
    chapters_lesser?: InputMaybe<Scalars['Int']['input']>;
    countryOfOrigin?: InputMaybe<Scalars['CountryCode']['input']>;
    duration?: InputMaybe<Scalars['Int']['input']>;
    duration_greater?: InputMaybe<Scalars['Int']['input']>;
    duration_lesser?: InputMaybe<Scalars['Int']['input']>;
    endDate?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    endDate_greater?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    endDate_lesser?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    endDate_like?: InputMaybe<Scalars['String']['input']>;
    episodes?: InputMaybe<Scalars['Int']['input']>;
    episodes_greater?: InputMaybe<Scalars['Int']['input']>;
    episodes_lesser?: InputMaybe<Scalars['Int']['input']>;
    format?: InputMaybe<MediaFormat>;
    format_in?: InputMaybe<Array<InputMaybe<MediaFormat>>>;
    format_not?: InputMaybe<MediaFormat>;
    format_not_in?: InputMaybe<Array<InputMaybe<MediaFormat>>>;
    genre?: InputMaybe<Scalars['String']['input']>;
    genre_in?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    genre_not_in?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    id?: InputMaybe<Scalars['Int']['input']>;
    idMal?: InputMaybe<Scalars['Int']['input']>;
    idMal_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    idMal_not?: InputMaybe<Scalars['Int']['input']>;
    idMal_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    id_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    id_not?: InputMaybe<Scalars['Int']['input']>;
    id_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    isAdult?: InputMaybe<Scalars['Boolean']['input']>;
    isLicensed?: InputMaybe<Scalars['Boolean']['input']>;
    licensedBy?: InputMaybe<Scalars['String']['input']>;
    licensedById?: InputMaybe<Scalars['Int']['input']>;
    licensedById_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    licensedBy_in?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    minimumTagRank?: InputMaybe<Scalars['Int']['input']>;
    onList?: InputMaybe<Scalars['Boolean']['input']>;
    popularity?: InputMaybe<Scalars['Int']['input']>;
    popularity_greater?: InputMaybe<Scalars['Int']['input']>;
    popularity_lesser?: InputMaybe<Scalars['Int']['input']>;
    popularity_not?: InputMaybe<Scalars['Int']['input']>;
    search?: InputMaybe<Scalars['String']['input']>;
    season?: InputMaybe<MediaSeason>;
    seasonYear?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<MediaSort>>>;
    source?: InputMaybe<MediaSource>;
    source_in?: InputMaybe<Array<InputMaybe<MediaSource>>>;
    startDate?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    startDate_greater?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    startDate_lesser?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    startDate_like?: InputMaybe<Scalars['String']['input']>;
    status?: InputMaybe<MediaStatus>;
    status_in?: InputMaybe<Array<InputMaybe<MediaStatus>>>;
    status_not?: InputMaybe<MediaStatus>;
    status_not_in?: InputMaybe<Array<InputMaybe<MediaStatus>>>;
    tag?: InputMaybe<Scalars['String']['input']>;
    tagCategory?: InputMaybe<Scalars['String']['input']>;
    tagCategory_in?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    tagCategory_not_in?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    tag_in?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    tag_not_in?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    type?: InputMaybe<MediaType>;
    volumes?: InputMaybe<Scalars['Int']['input']>;
    volumes_greater?: InputMaybe<Scalars['Int']['input']>;
    volumes_lesser?: InputMaybe<Scalars['Int']['input']>;
};


export type QueryMediaListArgs = {
    compareWithAuthList?: InputMaybe<Scalars['Boolean']['input']>;
    completedAt?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    completedAt_greater?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    completedAt_lesser?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    completedAt_like?: InputMaybe<Scalars['String']['input']>;
    id?: InputMaybe<Scalars['Int']['input']>;
    isFollowing?: InputMaybe<Scalars['Boolean']['input']>;
    mediaId?: InputMaybe<Scalars['Int']['input']>;
    mediaId_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    mediaId_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    notes?: InputMaybe<Scalars['String']['input']>;
    notes_like?: InputMaybe<Scalars['String']['input']>;
    sort?: InputMaybe<Array<InputMaybe<MediaListSort>>>;
    startedAt?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    startedAt_greater?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    startedAt_lesser?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    startedAt_like?: InputMaybe<Scalars['String']['input']>;
    status?: InputMaybe<MediaListStatus>;
    status_in?: InputMaybe<Array<InputMaybe<MediaListStatus>>>;
    status_not?: InputMaybe<MediaListStatus>;
    status_not_in?: InputMaybe<Array<InputMaybe<MediaListStatus>>>;
    type?: InputMaybe<MediaType>;
    userId?: InputMaybe<Scalars['Int']['input']>;
    userId_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    userName?: InputMaybe<Scalars['String']['input']>;
};


export type QueryMediaListCollectionArgs = {
    chunk?: InputMaybe<Scalars['Int']['input']>;
    completedAt?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    completedAt_greater?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    completedAt_lesser?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    completedAt_like?: InputMaybe<Scalars['String']['input']>;
    forceSingleCompletedList?: InputMaybe<Scalars['Boolean']['input']>;
    notes?: InputMaybe<Scalars['String']['input']>;
    notes_like?: InputMaybe<Scalars['String']['input']>;
    perChunk?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<MediaListSort>>>;
    startedAt?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    startedAt_greater?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    startedAt_lesser?: InputMaybe<Scalars['FuzzyDateInt']['input']>;
    startedAt_like?: InputMaybe<Scalars['String']['input']>;
    status?: InputMaybe<MediaListStatus>;
    status_in?: InputMaybe<Array<InputMaybe<MediaListStatus>>>;
    status_not?: InputMaybe<MediaListStatus>;
    status_not_in?: InputMaybe<Array<InputMaybe<MediaListStatus>>>;
    type?: InputMaybe<MediaType>;
    userId?: InputMaybe<Scalars['Int']['input']>;
    userName?: InputMaybe<Scalars['String']['input']>;
};


export type QueryMediaTagCollectionArgs = {
    status?: InputMaybe<Scalars['Int']['input']>;
};


export type QueryMediaTrendArgs = {
    averageScore?: InputMaybe<Scalars['Int']['input']>;
    averageScore_greater?: InputMaybe<Scalars['Int']['input']>;
    averageScore_lesser?: InputMaybe<Scalars['Int']['input']>;
    averageScore_not?: InputMaybe<Scalars['Int']['input']>;
    date?: InputMaybe<Scalars['Int']['input']>;
    date_greater?: InputMaybe<Scalars['Int']['input']>;
    date_lesser?: InputMaybe<Scalars['Int']['input']>;
    episode?: InputMaybe<Scalars['Int']['input']>;
    episode_greater?: InputMaybe<Scalars['Int']['input']>;
    episode_lesser?: InputMaybe<Scalars['Int']['input']>;
    episode_not?: InputMaybe<Scalars['Int']['input']>;
    mediaId?: InputMaybe<Scalars['Int']['input']>;
    mediaId_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    mediaId_not?: InputMaybe<Scalars['Int']['input']>;
    mediaId_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    popularity?: InputMaybe<Scalars['Int']['input']>;
    popularity_greater?: InputMaybe<Scalars['Int']['input']>;
    popularity_lesser?: InputMaybe<Scalars['Int']['input']>;
    popularity_not?: InputMaybe<Scalars['Int']['input']>;
    releasing?: InputMaybe<Scalars['Boolean']['input']>;
    sort?: InputMaybe<Array<InputMaybe<MediaTrendSort>>>;
    trending?: InputMaybe<Scalars['Int']['input']>;
    trending_greater?: InputMaybe<Scalars['Int']['input']>;
    trending_lesser?: InputMaybe<Scalars['Int']['input']>;
    trending_not?: InputMaybe<Scalars['Int']['input']>;
};


export type QueryNotificationArgs = {
    resetNotificationCount?: InputMaybe<Scalars['Boolean']['input']>;
    type?: InputMaybe<NotificationType>;
    type_in?: InputMaybe<Array<InputMaybe<NotificationType>>>;
};


export type QueryPageArgs = {
    page?: InputMaybe<Scalars['Int']['input']>;
    perPage?: InputMaybe<Scalars['Int']['input']>;
};


export type QueryRecommendationArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
    mediaId?: InputMaybe<Scalars['Int']['input']>;
    mediaRecommendationId?: InputMaybe<Scalars['Int']['input']>;
    onList?: InputMaybe<Scalars['Boolean']['input']>;
    rating?: InputMaybe<Scalars['Int']['input']>;
    rating_greater?: InputMaybe<Scalars['Int']['input']>;
    rating_lesser?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<RecommendationSort>>>;
    userId?: InputMaybe<Scalars['Int']['input']>;
};


export type QueryReviewArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
    mediaId?: InputMaybe<Scalars['Int']['input']>;
    mediaType?: InputMaybe<MediaType>;
    sort?: InputMaybe<Array<InputMaybe<ReviewSort>>>;
    userId?: InputMaybe<Scalars['Int']['input']>;
};


export type QueryStaffArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
    id_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    id_not?: InputMaybe<Scalars['Int']['input']>;
    id_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    isBirthday?: InputMaybe<Scalars['Boolean']['input']>;
    search?: InputMaybe<Scalars['String']['input']>;
    sort?: InputMaybe<Array<InputMaybe<StaffSort>>>;
};


export type QueryStudioArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
    id_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    id_not?: InputMaybe<Scalars['Int']['input']>;
    id_not_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    search?: InputMaybe<Scalars['String']['input']>;
    sort?: InputMaybe<Array<InputMaybe<StudioSort>>>;
};


export type QueryThreadArgs = {
    categoryId?: InputMaybe<Scalars['Int']['input']>;
    id?: InputMaybe<Scalars['Int']['input']>;
    id_in?: InputMaybe<Array<InputMaybe<Scalars['Int']['input']>>>;
    mediaCategoryId?: InputMaybe<Scalars['Int']['input']>;
    replyUserId?: InputMaybe<Scalars['Int']['input']>;
    search?: InputMaybe<Scalars['String']['input']>;
    sort?: InputMaybe<Array<InputMaybe<ThreadSort>>>;
    subscribed?: InputMaybe<Scalars['Boolean']['input']>;
    userId?: InputMaybe<Scalars['Int']['input']>;
};


export type QueryThreadCommentArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<ThreadCommentSort>>>;
    threadId?: InputMaybe<Scalars['Int']['input']>;
    userId?: InputMaybe<Scalars['Int']['input']>;
};


export type QueryUserArgs = {
    id?: InputMaybe<Scalars['Int']['input']>;
    isModerator?: InputMaybe<Scalars['Boolean']['input']>;
    name?: InputMaybe<Scalars['String']['input']>;
    search?: InputMaybe<Scalars['String']['input']>;
    sort?: InputMaybe<Array<InputMaybe<UserSort>>>;
};

/** Media recommendation */
export type Recommendation = {
    /** The id of the recommendation */
    id: Scalars['Int']['output'];
    /** The media the recommendation is from */
    media?: Maybe<Media>;
    /** The recommended media */
    mediaRecommendation?: Maybe<Media>;
    /** Users rating of the recommendation */
    rating?: Maybe<Scalars['Int']['output']>;
    /** The user that first created the recommendation */
    user?: Maybe<User>;
    /** The rating of the recommendation by currently authenticated user */
    userRating?: Maybe<RecommendationRating>;
};

export type RecommendationConnection = {
    edges?: Maybe<Array<Maybe<RecommendationEdge>>>;
    nodes?: Maybe<Array<Maybe<Recommendation>>>;
    /** The pagination information */
    pageInfo?: Maybe<PageInfo>;
};

/** Recommendation connection edge */
export type RecommendationEdge = {
    node?: Maybe<Recommendation>;
};

/** Recommendation rating enums */
export type RecommendationRating =
    | 'NO_RATING'
    | 'RATE_DOWN'
    | 'RATE_UP';

/** Recommendation sort enums */
export type RecommendationSort =
    | 'ID'
    | 'ID_DESC'
    | 'RATING'
    | 'RATING_DESC';

/** Notification for when new media is added to the site */
export type RelatedMediaAdditionNotification = {
    /** The notification context text */
    context?: Maybe<Scalars['String']['output']>;
    /** The time the notification was created at */
    createdAt?: Maybe<Scalars['Int']['output']>;
    /** The id of the Notification */
    id: Scalars['Int']['output'];
    /** The associated media of the airing schedule */
    media?: Maybe<Media>;
    /** The id of the new media */
    mediaId: Scalars['Int']['output'];
    /** The type of notification */
    type?: Maybe<NotificationType>;
};

export type Report = {
    cleared?: Maybe<Scalars['Boolean']['output']>;
    /** When the entry data was created */
    createdAt?: Maybe<Scalars['Int']['output']>;
    id: Scalars['Int']['output'];
    reason?: Maybe<Scalars['String']['output']>;
    reported?: Maybe<User>;
    reporter?: Maybe<User>;
};

/** A Review that features in an anime or manga */
export type Review = {
    /** The main review body text */
    body?: Maybe<Scalars['String']['output']>;
    /** The time of the thread creation */
    createdAt: Scalars['Int']['output'];
    /** The id of the review */
    id: Scalars['Int']['output'];
    /** The media the review is of */
    media?: Maybe<Media>;
    /** The id of the review's media */
    mediaId: Scalars['Int']['output'];
    /** For which type of media the review is for */
    mediaType?: Maybe<MediaType>;
    /** If the review is not yet publicly published and is only viewable by creator */
    private?: Maybe<Scalars['Boolean']['output']>;
    /** The total user rating of the review */
    rating?: Maybe<Scalars['Int']['output']>;
    /** The amount of user ratings of the review */
    ratingAmount?: Maybe<Scalars['Int']['output']>;
    /** The review score of the media */
    score?: Maybe<Scalars['Int']['output']>;
    /** The url for the review page on the AniList website */
    siteUrl?: Maybe<Scalars['String']['output']>;
    /** A short summary of the review */
    summary?: Maybe<Scalars['String']['output']>;
    /** The time of the thread last update */
    updatedAt: Scalars['Int']['output'];
    /** The creator of the review */
    user?: Maybe<User>;
    /** The id of the review's creator */
    userId: Scalars['Int']['output'];
    /** The rating of the review by currently authenticated user */
    userRating?: Maybe<ReviewRating>;
};


/** A Review that features in an anime or manga */
export type ReviewBodyArgs = {
    asHtml?: InputMaybe<Scalars['Boolean']['input']>;
};

export type ReviewConnection = {
    edges?: Maybe<Array<Maybe<ReviewEdge>>>;
    nodes?: Maybe<Array<Maybe<Review>>>;
    /** The pagination information */
    pageInfo?: Maybe<PageInfo>;
};

/** Review connection edge */
export type ReviewEdge = {
    node?: Maybe<Review>;
};

/** Review rating enums */
export type ReviewRating =
    | 'DOWN_VOTE'
    | 'NO_VOTE'
    | 'UP_VOTE';

/** Review sort enums */
export type ReviewSort =
    | 'CREATED_AT'
    | 'CREATED_AT_DESC'
    | 'ID'
    | 'ID_DESC'
    | 'RATING'
    | 'RATING_DESC'
    | 'SCORE'
    | 'SCORE_DESC'
    | 'UPDATED_AT'
    | 'UPDATED_AT_DESC';

/** Feed of mod edit activity */
export type RevisionHistory = {
    /** The action taken on the objects */
    action?: Maybe<RevisionHistoryAction>;
    /** A JSON object of the fields that changed */
    changes?: Maybe<Scalars['Json']['output']>;
    /** The character the mod feed entry references */
    character?: Maybe<Character>;
    /** When the mod feed entry was created */
    createdAt?: Maybe<Scalars['Int']['output']>;
    /** The external link source the mod feed entry references */
    externalLink?: Maybe<MediaExternalLink>;
    /** The id of the media */
    id: Scalars['Int']['output'];
    /** The media the mod feed entry references */
    media?: Maybe<Media>;
    /** The staff member the mod feed entry references */
    staff?: Maybe<Staff>;
    /** The studio the mod feed entry references */
    studio?: Maybe<Studio>;
    /** The user who made the edit to the object */
    user?: Maybe<User>;
};

/** Revision history actions */
export type RevisionHistoryAction =
    | 'CREATE'
    | 'EDIT';

/** A user's list score distribution. */
export type ScoreDistribution = {
    /** The amount of list entries with this score */
    amount?: Maybe<Scalars['Int']['output']>;
    score?: Maybe<Scalars['Int']['output']>;
};

/** Media list scoring type */
export type ScoreFormat =
/** An integer from 0-3. Should be represented in Smileys. 0 => No Score, 1 => :(, 2 => :|, 3 => :) */
    | 'POINT_3'
    /** An integer from 0-5. Should be represented in Stars */
    | 'POINT_5'
    /** An integer from 0-10 */
    | 'POINT_10'
    /** A float from 0-10 with 1 decimal place */
    | 'POINT_10_DECIMAL'
    /** An integer from 0-100 */
    | 'POINT_100';

export type SiteStatistics = {
    anime?: Maybe<SiteTrendConnection>;
    characters?: Maybe<SiteTrendConnection>;
    manga?: Maybe<SiteTrendConnection>;
    reviews?: Maybe<SiteTrendConnection>;
    staff?: Maybe<SiteTrendConnection>;
    studios?: Maybe<SiteTrendConnection>;
    users?: Maybe<SiteTrendConnection>;
};


export type SiteStatisticsAnimeArgs = {
    page?: InputMaybe<Scalars['Int']['input']>;
    perPage?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<SiteTrendSort>>>;
};


export type SiteStatisticsCharactersArgs = {
    page?: InputMaybe<Scalars['Int']['input']>;
    perPage?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<SiteTrendSort>>>;
};


export type SiteStatisticsMangaArgs = {
    page?: InputMaybe<Scalars['Int']['input']>;
    perPage?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<SiteTrendSort>>>;
};


export type SiteStatisticsReviewsArgs = {
    page?: InputMaybe<Scalars['Int']['input']>;
    perPage?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<SiteTrendSort>>>;
};


export type SiteStatisticsStaffArgs = {
    page?: InputMaybe<Scalars['Int']['input']>;
    perPage?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<SiteTrendSort>>>;
};


export type SiteStatisticsStudiosArgs = {
    page?: InputMaybe<Scalars['Int']['input']>;
    perPage?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<SiteTrendSort>>>;
};


export type SiteStatisticsUsersArgs = {
    page?: InputMaybe<Scalars['Int']['input']>;
    perPage?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<SiteTrendSort>>>;
};

/** Daily site statistics */
export type SiteTrend = {
    /** The change from yesterday */
    change: Scalars['Int']['output'];
    count: Scalars['Int']['output'];
    /** The day the data was recorded (timestamp) */
    date: Scalars['Int']['output'];
};

export type SiteTrendConnection = {
    edges?: Maybe<Array<Maybe<SiteTrendEdge>>>;
    nodes?: Maybe<Array<Maybe<SiteTrend>>>;
    /** The pagination information */
    pageInfo?: Maybe<PageInfo>;
};

/** Site trend connection edge */
export type SiteTrendEdge = {
    node?: Maybe<SiteTrend>;
};

/** Site trend sort enums */
export type SiteTrendSort =
    | 'CHANGE'
    | 'CHANGE_DESC'
    | 'COUNT'
    | 'COUNT_DESC'
    | 'DATE'
    | 'DATE_DESC';

/** Voice actors or production staff */
export type Staff = {
    /** The person's age in years */
    age?: Maybe<Scalars['Int']['output']>;
    /** The persons blood type */
    bloodType?: Maybe<Scalars['String']['output']>;
    /** Media the actor voiced characters in. (Same data as characters with media as node instead of characters) */
    characterMedia?: Maybe<MediaConnection>;
    /** Characters voiced by the actor */
    characters?: Maybe<CharacterConnection>;
    dateOfBirth?: Maybe<FuzzyDate>;
    dateOfDeath?: Maybe<FuzzyDate>;
    /** A general description of the staff member */
    description?: Maybe<Scalars['String']['output']>;
    /** The amount of user's who have favourited the staff member */
    favourites?: Maybe<Scalars['Int']['output']>;
    /** The staff's gender. Usually Male, Female, or Non-binary but can be any string. */
    gender?: Maybe<Scalars['String']['output']>;
    /** The persons birthplace or hometown */
    homeTown?: Maybe<Scalars['String']['output']>;
    /** The id of the staff member */
    id: Scalars['Int']['output'];
    /** The staff images */
    image?: Maybe<StaffImage>;
    /** If the staff member is marked as favourite by the currently authenticated user */
    isFavourite: Scalars['Boolean']['output'];
    /** If the staff member is blocked from being added to favourites */
    isFavouriteBlocked: Scalars['Boolean']['output'];
    /**
     * The primary language the staff member dub's in
     * @deprecated Replaced with languageV2
     */
    language?: Maybe<StaffLanguage>;
    /** The primary language of the staff member. Current values: Japanese, English, Korean, Italian, Spanish, Portuguese, French, German, Hebrew, Hungarian, Chinese, Arabic, Filipino, Catalan, Finnish, Turkish, Dutch, Swedish, Thai, Tagalog, Malaysian, Indonesian, Vietnamese, Nepali, Hindi, Urdu */
    languageV2?: Maybe<Scalars['String']['output']>;
    /** Notes for site moderators */
    modNotes?: Maybe<Scalars['String']['output']>;
    /** The names of the staff member */
    name?: Maybe<StaffName>;
    /** The person's primary occupations */
    primaryOccupations?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
    /** The url for the staff page on the AniList website */
    siteUrl?: Maybe<Scalars['String']['output']>;
    /** Staff member that the submission is referencing */
    staff?: Maybe<Staff>;
    /** Media where the staff member has a production role */
    staffMedia?: Maybe<MediaConnection>;
    /** Inner details of submission status */
    submissionNotes?: Maybe<Scalars['String']['output']>;
    /** Status of the submission */
    submissionStatus?: Maybe<Scalars['Int']['output']>;
    /** Submitter for the submission */
    submitter?: Maybe<User>;
    /** @deprecated No data available */
    updatedAt?: Maybe<Scalars['Int']['output']>;
    /** [startYear, endYear] (If the 2nd value is not present staff is still active) */
    yearsActive?: Maybe<Array<Maybe<Scalars['Int']['output']>>>;
};


/** Voice actors or production staff */
export type StaffCharacterMediaArgs = {
    onList?: InputMaybe<Scalars['Boolean']['input']>;
    page?: InputMaybe<Scalars['Int']['input']>;
    perPage?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<MediaSort>>>;
};


/** Voice actors or production staff */
export type StaffCharactersArgs = {
    page?: InputMaybe<Scalars['Int']['input']>;
    perPage?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<CharacterSort>>>;
};


/** Voice actors or production staff */
export type StaffDescriptionArgs = {
    asHtml?: InputMaybe<Scalars['Boolean']['input']>;
};


/** Voice actors or production staff */
export type StaffStaffMediaArgs = {
    onList?: InputMaybe<Scalars['Boolean']['input']>;
    page?: InputMaybe<Scalars['Int']['input']>;
    perPage?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<MediaSort>>>;
    type?: InputMaybe<MediaType>;
};

export type StaffConnection = {
    edges?: Maybe<Array<Maybe<StaffEdge>>>;
    nodes?: Maybe<Array<Maybe<Staff>>>;
    /** The pagination information */
    pageInfo?: Maybe<PageInfo>;
};

/** Staff connection edge */
export type StaffEdge = {
    /** The order the staff should be displayed from the users favourites */
    favouriteOrder?: Maybe<Scalars['Int']['output']>;
    /** The id of the connection */
    id?: Maybe<Scalars['Int']['output']>;
    node?: Maybe<Staff>;
    /** The role of the staff member in the production of the media */
    role?: Maybe<Scalars['String']['output']>;
};

export type StaffImage = {
    /** The person's image of media at its largest size */
    large?: Maybe<Scalars['String']['output']>;
    /** The person's image of media at medium size */
    medium?: Maybe<Scalars['String']['output']>;
};

/** The primary language of the voice actor */
export type StaffLanguage =
/** English */
    | 'ENGLISH'
    /** French */
    | 'FRENCH'
    /** German */
    | 'GERMAN'
    /** Hebrew */
    | 'HEBREW'
    /** Hungarian */
    | 'HUNGARIAN'
    /** Italian */
    | 'ITALIAN'
    /** Japanese */
    | 'JAPANESE'
    /** Korean */
    | 'KOREAN'
    /** Portuguese */
    | 'PORTUGUESE'
    /** Spanish */
    | 'SPANISH';

/** The names of the staff member */
export type StaffName = {
    /** Other names the staff member might be referred to as (pen names) */
    alternative?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
    /** The person's given name */
    first?: Maybe<Scalars['String']['output']>;
    /** The person's first and last name */
    full?: Maybe<Scalars['String']['output']>;
    /** The person's surname */
    last?: Maybe<Scalars['String']['output']>;
    /** The person's middle name */
    middle?: Maybe<Scalars['String']['output']>;
    /** The person's full name in their native language */
    native?: Maybe<Scalars['String']['output']>;
    /** The currently authenticated users preferred name language. Default romaji for non-authenticated */
    userPreferred?: Maybe<Scalars['String']['output']>;
};

/** The names of the staff member */
export type StaffNameInput = {
    /** Other names the character might be referred by */
    alternative?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    /** The person's given name */
    first?: InputMaybe<Scalars['String']['input']>;
    /** The person's surname */
    last?: InputMaybe<Scalars['String']['input']>;
    /** The person's middle name */
    middle?: InputMaybe<Scalars['String']['input']>;
    /** The person's full name in their native language */
    native?: InputMaybe<Scalars['String']['input']>;
};

/** Voice actor role for a character */
export type StaffRoleType = {
    /** Used for grouping roles where multiple dubs exist for the same language. Either dubbing company name or language variant. */
    dubGroup?: Maybe<Scalars['String']['output']>;
    /** Notes regarding the VA's role for the character */
    roleNotes?: Maybe<Scalars['String']['output']>;
    /** The voice actors of the character */
    voiceActor?: Maybe<Staff>;
};

/** Staff sort enums */
export type StaffSort =
    | 'FAVOURITES'
    | 'FAVOURITES_DESC'
    | 'ID'
    | 'ID_DESC'
    | 'LANGUAGE'
    | 'LANGUAGE_DESC'
    /** Order manually decided by moderators */
    | 'RELEVANCE'
    | 'ROLE'
    | 'ROLE_DESC'
    | 'SEARCH_MATCH';

/** User's staff statistics */
export type StaffStats = {
    amount?: Maybe<Scalars['Int']['output']>;
    meanScore?: Maybe<Scalars['Int']['output']>;
    staff?: Maybe<Staff>;
    /** The amount of time in minutes the staff member has been watched by the user */
    timeWatched?: Maybe<Scalars['Int']['output']>;
};

/** A submission for a staff that features in an anime or manga */
export type StaffSubmission = {
    /** Data Mod assigned to handle the submission */
    assignee?: Maybe<User>;
    createdAt?: Maybe<Scalars['Int']['output']>;
    /** The id of the submission */
    id: Scalars['Int']['output'];
    /** Whether the submission is locked */
    locked?: Maybe<Scalars['Boolean']['output']>;
    /** Inner details of submission status */
    notes?: Maybe<Scalars['String']['output']>;
    source?: Maybe<Scalars['String']['output']>;
    /** Staff that the submission is referencing */
    staff?: Maybe<Staff>;
    /** Status of the submission */
    status?: Maybe<SubmissionStatus>;
    /** The staff submission changes */
    submission?: Maybe<Staff>;
    /** Submitter for the submission */
    submitter?: Maybe<User>;
};

/** The distribution of the watching/reading status of media or a user's list */
export type StatusDistribution = {
    /** The amount of entries with this status */
    amount?: Maybe<Scalars['Int']['output']>;
    /** The day the activity took place (Unix timestamp) */
    status?: Maybe<MediaListStatus>;
};

/** Animation or production company */
export type Studio = {
    /** The amount of user's who have favourited the studio */
    favourites?: Maybe<Scalars['Int']['output']>;
    /** The id of the studio */
    id: Scalars['Int']['output'];
    /** If the studio is an animation studio or a different kind of company */
    isAnimationStudio: Scalars['Boolean']['output'];
    /** If the studio is marked as favourite by the currently authenticated user */
    isFavourite: Scalars['Boolean']['output'];
    /** The media the studio has worked on */
    media?: Maybe<MediaConnection>;
    /** The name of the studio */
    name: Scalars['String']['output'];
    /** The url for the studio page on the AniList website */
    siteUrl?: Maybe<Scalars['String']['output']>;
};


/** Animation or production company */
export type StudioMediaArgs = {
    isMain?: InputMaybe<Scalars['Boolean']['input']>;
    onList?: InputMaybe<Scalars['Boolean']['input']>;
    page?: InputMaybe<Scalars['Int']['input']>;
    perPage?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<MediaSort>>>;
};

export type StudioConnection = {
    edges?: Maybe<Array<Maybe<StudioEdge>>>;
    nodes?: Maybe<Array<Maybe<Studio>>>;
    /** The pagination information */
    pageInfo?: Maybe<PageInfo>;
};

/** Studio connection edge */
export type StudioEdge = {
    /** The order the character should be displayed from the users favourites */
    favouriteOrder?: Maybe<Scalars['Int']['output']>;
    /** The id of the connection */
    id?: Maybe<Scalars['Int']['output']>;
    /** If the studio is the main animation studio of the anime */
    isMain: Scalars['Boolean']['output'];
    node?: Maybe<Studio>;
};

/** Studio sort enums */
export type StudioSort =
    | 'FAVOURITES'
    | 'FAVOURITES_DESC'
    | 'ID'
    | 'ID_DESC'
    | 'NAME'
    | 'NAME_DESC'
    | 'SEARCH_MATCH';

/** User's studio statistics */
export type StudioStats = {
    amount?: Maybe<Scalars['Int']['output']>;
    meanScore?: Maybe<Scalars['Int']['output']>;
    studio?: Maybe<Studio>;
    /** The amount of time in minutes the studio's works have been watched by the user */
    timeWatched?: Maybe<Scalars['Int']['output']>;
};

/** Submission sort enums */
export type SubmissionSort =
    | 'ID'
    | 'ID_DESC';

/** Submission status */
export type SubmissionStatus =
    | 'ACCEPTED'
    | 'PARTIALLY_ACCEPTED'
    | 'PENDING'
    | 'REJECTED';

/** User's tag statistics */
export type TagStats = {
    amount?: Maybe<Scalars['Int']['output']>;
    meanScore?: Maybe<Scalars['Int']['output']>;
    tag?: Maybe<MediaTag>;
    /** The amount of time in minutes the tag has been watched by the user */
    timeWatched?: Maybe<Scalars['Int']['output']>;
};

/** User text activity */
export type TextActivity = {
    /** The time the activity was created at */
    createdAt: Scalars['Int']['output'];
    /** The id of the activity */
    id: Scalars['Int']['output'];
    /** If the currently authenticated user liked the activity */
    isLiked?: Maybe<Scalars['Boolean']['output']>;
    /** If the activity is locked and can receive replies */
    isLocked?: Maybe<Scalars['Boolean']['output']>;
    /** If the activity is pinned to the top of the users activity feed */
    isPinned?: Maybe<Scalars['Boolean']['output']>;
    /** If the currently authenticated user is subscribed to the activity */
    isSubscribed?: Maybe<Scalars['Boolean']['output']>;
    /** The amount of likes the activity has */
    likeCount: Scalars['Int']['output'];
    /** The users who liked the activity */
    likes?: Maybe<Array<Maybe<User>>>;
    /** The written replies to the activity */
    replies?: Maybe<Array<Maybe<ActivityReply>>>;
    /** The number of activity replies */
    replyCount: Scalars['Int']['output'];
    /** The url for the activity page on the AniList website */
    siteUrl?: Maybe<Scalars['String']['output']>;
    /** The status text (Markdown) */
    text?: Maybe<Scalars['String']['output']>;
    /** The type of activity */
    type?: Maybe<ActivityType>;
    /** The user who created the activity */
    user?: Maybe<User>;
    /** The user id of the activity's creator */
    userId?: Maybe<Scalars['Int']['output']>;
};


/** User text activity */
export type TextActivityTextArgs = {
    asHtml?: InputMaybe<Scalars['Boolean']['input']>;
};

/** Forum Thread */
export type Thread = {
    /** The text body of the thread (Markdown) */
    body?: Maybe<Scalars['String']['output']>;
    /** The categories of the thread */
    categories?: Maybe<Array<Maybe<ThreadCategory>>>;
    /** The time of the thread creation */
    createdAt: Scalars['Int']['output'];
    /** The id of the thread */
    id: Scalars['Int']['output'];
    /** If the currently authenticated user liked the thread */
    isLiked?: Maybe<Scalars['Boolean']['output']>;
    /** If the thread is locked and can receive comments */
    isLocked?: Maybe<Scalars['Boolean']['output']>;
    /** If the thread is stickied and should be displayed at the top of the page */
    isSticky?: Maybe<Scalars['Boolean']['output']>;
    /** If the currently authenticated user is subscribed to the thread */
    isSubscribed?: Maybe<Scalars['Boolean']['output']>;
    /** The amount of likes the thread has */
    likeCount: Scalars['Int']['output'];
    /** The users who liked the thread */
    likes?: Maybe<Array<Maybe<User>>>;
    /** The media categories of the thread */
    mediaCategories?: Maybe<Array<Maybe<Media>>>;
    /** The time of the last reply */
    repliedAt?: Maybe<Scalars['Int']['output']>;
    /** The id of the most recent comment on the thread */
    replyCommentId?: Maybe<Scalars['Int']['output']>;
    /** The number of comments on the thread */
    replyCount?: Maybe<Scalars['Int']['output']>;
    /** The user to last reply to the thread */
    replyUser?: Maybe<User>;
    /** The id of the user who most recently commented on the thread */
    replyUserId?: Maybe<Scalars['Int']['output']>;
    /** The url for the thread page on the AniList website */
    siteUrl?: Maybe<Scalars['String']['output']>;
    /** The title of the thread */
    title?: Maybe<Scalars['String']['output']>;
    /** The time of the thread last update */
    updatedAt: Scalars['Int']['output'];
    /** The owner of the thread */
    user?: Maybe<User>;
    /** The id of the thread owner user */
    userId: Scalars['Int']['output'];
    /** The number of times users have viewed the thread */
    viewCount?: Maybe<Scalars['Int']['output']>;
};


/** Forum Thread */
export type ThreadBodyArgs = {
    asHtml?: InputMaybe<Scalars['Boolean']['input']>;
};

/** A forum thread category */
export type ThreadCategory = {
    /** The id of the category */
    id: Scalars['Int']['output'];
    /** The name of the category */
    name: Scalars['String']['output'];
};

/** Forum Thread Comment */
export type ThreadComment = {
    /** The comment's child reply comments */
    childComments?: Maybe<Scalars['Json']['output']>;
    /** The text content of the comment (Markdown) */
    comment?: Maybe<Scalars['String']['output']>;
    /** The time of the comments creation */
    createdAt: Scalars['Int']['output'];
    /** The id of the comment */
    id: Scalars['Int']['output'];
    /** If the currently authenticated user liked the comment */
    isLiked?: Maybe<Scalars['Boolean']['output']>;
    /** If the comment tree is locked and may not receive replies or edits */
    isLocked?: Maybe<Scalars['Boolean']['output']>;
    /** The amount of likes the comment has */
    likeCount: Scalars['Int']['output'];
    /** The users who liked the comment */
    likes?: Maybe<Array<Maybe<User>>>;
    /** The url for the comment page on the AniList website */
    siteUrl?: Maybe<Scalars['String']['output']>;
    /** The thread the comment belongs to */
    thread?: Maybe<Thread>;
    /** The id of thread the comment belongs to */
    threadId?: Maybe<Scalars['Int']['output']>;
    /** The time of the comments last update */
    updatedAt: Scalars['Int']['output'];
    /** The user who created the comment */
    user?: Maybe<User>;
    /** The user id of the comment's owner */
    userId?: Maybe<Scalars['Int']['output']>;
};


/** Forum Thread Comment */
export type ThreadCommentCommentArgs = {
    asHtml?: InputMaybe<Scalars['Boolean']['input']>;
};

/** Notification for when a thread comment is liked */
export type ThreadCommentLikeNotification = {
    /** The thread comment that was liked */
    comment?: Maybe<ThreadComment>;
    /** The id of the activity which was liked */
    commentId: Scalars['Int']['output'];
    /** The notification context text */
    context?: Maybe<Scalars['String']['output']>;
    /** The time the notification was created at */
    createdAt?: Maybe<Scalars['Int']['output']>;
    /** The id of the Notification */
    id: Scalars['Int']['output'];
    /** The thread that the relevant comment belongs to */
    thread?: Maybe<Thread>;
    /** The type of notification */
    type?: Maybe<NotificationType>;
    /** The user who liked the activity */
    user?: Maybe<User>;
    /** The id of the user who liked to the activity */
    userId: Scalars['Int']['output'];
};

/** Notification for when authenticated user is @ mentioned in a forum thread comment */
export type ThreadCommentMentionNotification = {
    /** The thread comment that included the @ mention */
    comment?: Maybe<ThreadComment>;
    /** The id of the comment where mentioned */
    commentId: Scalars['Int']['output'];
    /** The notification context text */
    context?: Maybe<Scalars['String']['output']>;
    /** The time the notification was created at */
    createdAt?: Maybe<Scalars['Int']['output']>;
    /** The id of the Notification */
    id: Scalars['Int']['output'];
    /** The thread that the relevant comment belongs to */
    thread?: Maybe<Thread>;
    /** The type of notification */
    type?: Maybe<NotificationType>;
    /** The user who mentioned the authenticated user */
    user?: Maybe<User>;
    /** The id of the user who mentioned the authenticated user */
    userId: Scalars['Int']['output'];
};

/** Notification for when a user replies to your forum thread comment */
export type ThreadCommentReplyNotification = {
    /** The reply thread comment */
    comment?: Maybe<ThreadComment>;
    /** The id of the reply comment */
    commentId: Scalars['Int']['output'];
    /** The notification context text */
    context?: Maybe<Scalars['String']['output']>;
    /** The time the notification was created at */
    createdAt?: Maybe<Scalars['Int']['output']>;
    /** The id of the Notification */
    id: Scalars['Int']['output'];
    /** The thread that the relevant comment belongs to */
    thread?: Maybe<Thread>;
    /** The type of notification */
    type?: Maybe<NotificationType>;
    /** The user who replied to the activity */
    user?: Maybe<User>;
    /** The id of the user who create the comment reply */
    userId: Scalars['Int']['output'];
};

/** Thread comments sort enums */
export type ThreadCommentSort =
    | 'ID'
    | 'ID_DESC';

/** Notification for when a user replies to a subscribed forum thread */
export type ThreadCommentSubscribedNotification = {
    /** The reply thread comment */
    comment?: Maybe<ThreadComment>;
    /** The id of the new comment in the subscribed thread */
    commentId: Scalars['Int']['output'];
    /** The notification context text */
    context?: Maybe<Scalars['String']['output']>;
    /** The time the notification was created at */
    createdAt?: Maybe<Scalars['Int']['output']>;
    /** The id of the Notification */
    id: Scalars['Int']['output'];
    /** The thread that the relevant comment belongs to */
    thread?: Maybe<Thread>;
    /** The type of notification */
    type?: Maybe<NotificationType>;
    /** The user who replied to the subscribed thread */
    user?: Maybe<User>;
    /** The id of the user who commented on the thread */
    userId: Scalars['Int']['output'];
};

/** Notification for when a thread is liked */
export type ThreadLikeNotification = {
    /** The liked thread comment */
    comment?: Maybe<ThreadComment>;
    /** The notification context text */
    context?: Maybe<Scalars['String']['output']>;
    /** The time the notification was created at */
    createdAt?: Maybe<Scalars['Int']['output']>;
    /** The id of the Notification */
    id: Scalars['Int']['output'];
    /** The thread that the relevant comment belongs to */
    thread?: Maybe<Thread>;
    /** The id of the thread which was liked */
    threadId: Scalars['Int']['output'];
    /** The type of notification */
    type?: Maybe<NotificationType>;
    /** The user who liked the activity */
    user?: Maybe<User>;
    /** The id of the user who liked to the activity */
    userId: Scalars['Int']['output'];
};

/** Thread sort enums */
export type ThreadSort =
    | 'CREATED_AT'
    | 'CREATED_AT_DESC'
    | 'ID'
    | 'ID_DESC'
    | 'IS_STICKY'
    | 'REPLIED_AT'
    | 'REPLIED_AT_DESC'
    | 'REPLY_COUNT'
    | 'REPLY_COUNT_DESC'
    | 'SEARCH_MATCH'
    | 'TITLE'
    | 'TITLE_DESC'
    | 'UPDATED_AT'
    | 'UPDATED_AT_DESC'
    | 'VIEW_COUNT'
    | 'VIEW_COUNT_DESC';

/** A user */
export type User = {
    /** The bio written by user (Markdown) */
    about?: Maybe<Scalars['String']['output']>;
    /** The user's avatar images */
    avatar?: Maybe<UserAvatar>;
    /** The user's banner images */
    bannerImage?: Maybe<Scalars['String']['output']>;
    bans?: Maybe<Scalars['Json']['output']>;
    /** When the user's account was created. (Does not exist for accounts created before 2020) */
    createdAt?: Maybe<Scalars['Int']['output']>;
    /** Custom donation badge text */
    donatorBadge?: Maybe<Scalars['String']['output']>;
    /** The donation tier of the user */
    donatorTier?: Maybe<Scalars['Int']['output']>;
    /** The users favourites */
    favourites?: Maybe<Favourites>;
    /** The id of the user */
    id: Scalars['Int']['output'];
    /** If the user is blocked by the authenticated user */
    isBlocked?: Maybe<Scalars['Boolean']['output']>;
    /** If this user if following the authenticated user */
    isFollower?: Maybe<Scalars['Boolean']['output']>;
    /** If the authenticated user if following this user */
    isFollowing?: Maybe<Scalars['Boolean']['output']>;
    /** The user's media list options */
    mediaListOptions?: Maybe<MediaListOptions>;
    /** The user's moderator roles if they are a site moderator */
    moderatorRoles?: Maybe<Array<Maybe<ModRole>>>;
    /**
     * If the user is a moderator or data moderator
     * @deprecated Deprecated. Replaced with moderatorRoles field.
     */
    moderatorStatus?: Maybe<Scalars['String']['output']>;
    /** The name of the user */
    name: Scalars['String']['output'];
    /** The user's general options */
    options?: Maybe<UserOptions>;
    /** The user's previously used names. */
    previousNames?: Maybe<Array<Maybe<UserPreviousName>>>;
    /** The url for the user page on the AniList website */
    siteUrl?: Maybe<Scalars['String']['output']>;
    /** The users anime & manga list statistics */
    statistics?: Maybe<UserStatisticTypes>;
    /**
     * The user's statistics
     * @deprecated Deprecated. Replaced with statistics field.
     */
    stats?: Maybe<UserStats>;
    /** The number of unread notifications the user has */
    unreadNotificationCount?: Maybe<Scalars['Int']['output']>;
    /** When the user's data was last updated */
    updatedAt?: Maybe<Scalars['Int']['output']>;
};


/** A user */
export type UserAboutArgs = {
    asHtml?: InputMaybe<Scalars['Boolean']['input']>;
};


/** A user */
export type UserFavouritesArgs = {
    page?: InputMaybe<Scalars['Int']['input']>;
};

/** A user's activity history stats. */
export type UserActivityHistory = {
    /** The amount of activity on the day */
    amount?: Maybe<Scalars['Int']['output']>;
    /** The day the activity took place (Unix timestamp) */
    date?: Maybe<Scalars['Int']['output']>;
    /** The level of activity represented on a 1-10 scale */
    level?: Maybe<Scalars['Int']['output']>;
};

/** A user's avatars */
export type UserAvatar = {
    /** The avatar of user at its largest size */
    large?: Maybe<Scalars['String']['output']>;
    /** The avatar of user at medium size */
    medium?: Maybe<Scalars['String']['output']>;
};

export type UserCountryStatistic = {
    chaptersRead: Scalars['Int']['output'];
    count: Scalars['Int']['output'];
    country?: Maybe<Scalars['CountryCode']['output']>;
    meanScore: Scalars['Float']['output'];
    mediaIds: Array<Maybe<Scalars['Int']['output']>>;
    minutesWatched: Scalars['Int']['output'];
};

export type UserFormatStatistic = {
    chaptersRead: Scalars['Int']['output'];
    count: Scalars['Int']['output'];
    format?: Maybe<MediaFormat>;
    meanScore: Scalars['Float']['output'];
    mediaIds: Array<Maybe<Scalars['Int']['output']>>;
    minutesWatched: Scalars['Int']['output'];
};

export type UserGenreStatistic = {
    chaptersRead: Scalars['Int']['output'];
    count: Scalars['Int']['output'];
    genre?: Maybe<Scalars['String']['output']>;
    meanScore: Scalars['Float']['output'];
    mediaIds: Array<Maybe<Scalars['Int']['output']>>;
    minutesWatched: Scalars['Int']['output'];
};

export type UserLengthStatistic = {
    chaptersRead: Scalars['Int']['output'];
    count: Scalars['Int']['output'];
    length?: Maybe<Scalars['String']['output']>;
    meanScore: Scalars['Float']['output'];
    mediaIds: Array<Maybe<Scalars['Int']['output']>>;
    minutesWatched: Scalars['Int']['output'];
};

/** User data for moderators */
export type UserModData = {
    alts?: Maybe<Array<Maybe<User>>>;
    bans?: Maybe<Scalars['Json']['output']>;
    counts?: Maybe<Scalars['Json']['output']>;
    email?: Maybe<Scalars['String']['output']>;
    ip?: Maybe<Scalars['Json']['output']>;
    privacy?: Maybe<Scalars['Int']['output']>;
};

/** A user's general options */
export type UserOptions = {
    /** Minutes between activity for them to be merged together. 0 is Never, Above 2 weeks (20160 mins) is Always. */
    activityMergeTime?: Maybe<Scalars['Int']['output']>;
    /** Whether the user receives notifications when a show they are watching aires */
    airingNotifications?: Maybe<Scalars['Boolean']['output']>;
    /** The list activity types the user has disabled from being created from list updates */
    disabledListActivity?: Maybe<Array<Maybe<ListActivityOption>>>;
    /** Whether the user has enabled viewing of 18+ content */
    displayAdultContent?: Maybe<Scalars['Boolean']['output']>;
    /** Notification options */
    notificationOptions?: Maybe<Array<Maybe<NotificationOption>>>;
    /** Profile highlight color (blue, purple, pink, orange, red, green, gray) */
    profileColor?: Maybe<Scalars['String']['output']>;
    /** Whether the user only allow messages from users they follow */
    restrictMessagesToFollowing?: Maybe<Scalars['Boolean']['output']>;
    /** The language the user wants to see staff and character names in */
    staffNameLanguage?: Maybe<UserStaffNameLanguage>;
    /** The user's timezone offset (Auth user only) */
    timezone?: Maybe<Scalars['String']['output']>;
    /** The language the user wants to see media titles in */
    titleLanguage?: Maybe<UserTitleLanguage>;
};

/** A user's previous name */
export type UserPreviousName = {
    /** When the user first changed from this name. */
    createdAt?: Maybe<Scalars['Int']['output']>;
    /** A previous name of the user. */
    name?: Maybe<Scalars['String']['output']>;
    /** When the user most recently changed from this name. */
    updatedAt?: Maybe<Scalars['Int']['output']>;
};

export type UserReleaseYearStatistic = {
    chaptersRead: Scalars['Int']['output'];
    count: Scalars['Int']['output'];
    meanScore: Scalars['Float']['output'];
    mediaIds: Array<Maybe<Scalars['Int']['output']>>;
    minutesWatched: Scalars['Int']['output'];
    releaseYear?: Maybe<Scalars['Int']['output']>;
};

export type UserScoreStatistic = {
    chaptersRead: Scalars['Int']['output'];
    count: Scalars['Int']['output'];
    meanScore: Scalars['Float']['output'];
    mediaIds: Array<Maybe<Scalars['Int']['output']>>;
    minutesWatched: Scalars['Int']['output'];
    score?: Maybe<Scalars['Int']['output']>;
};

/** User sort enums */
export type UserSort =
    | 'CHAPTERS_READ'
    | 'CHAPTERS_READ_DESC'
    | 'ID'
    | 'ID_DESC'
    | 'SEARCH_MATCH'
    | 'USERNAME'
    | 'USERNAME_DESC'
    | 'WATCHED_TIME'
    | 'WATCHED_TIME_DESC';

/** The language the user wants to see staff and character names in */
export type UserStaffNameLanguage =
/** The staff or character's name in their native language */
    | 'NATIVE'
    /** The romanization of the staff or character's native name */
    | 'ROMAJI'
    /** The romanization of the staff or character's native name, with western name ordering */
    | 'ROMAJI_WESTERN';

export type UserStaffStatistic = {
    chaptersRead: Scalars['Int']['output'];
    count: Scalars['Int']['output'];
    meanScore: Scalars['Float']['output'];
    mediaIds: Array<Maybe<Scalars['Int']['output']>>;
    minutesWatched: Scalars['Int']['output'];
    staff?: Maybe<Staff>;
};

export type UserStartYearStatistic = {
    chaptersRead: Scalars['Int']['output'];
    count: Scalars['Int']['output'];
    meanScore: Scalars['Float']['output'];
    mediaIds: Array<Maybe<Scalars['Int']['output']>>;
    minutesWatched: Scalars['Int']['output'];
    startYear?: Maybe<Scalars['Int']['output']>;
};

export type UserStatisticTypes = {
    anime?: Maybe<UserStatistics>;
    manga?: Maybe<UserStatistics>;
};

export type UserStatistics = {
    chaptersRead: Scalars['Int']['output'];
    count: Scalars['Int']['output'];
    countries?: Maybe<Array<Maybe<UserCountryStatistic>>>;
    episodesWatched: Scalars['Int']['output'];
    formats?: Maybe<Array<Maybe<UserFormatStatistic>>>;
    genres?: Maybe<Array<Maybe<UserGenreStatistic>>>;
    lengths?: Maybe<Array<Maybe<UserLengthStatistic>>>;
    meanScore: Scalars['Float']['output'];
    minutesWatched: Scalars['Int']['output'];
    releaseYears?: Maybe<Array<Maybe<UserReleaseYearStatistic>>>;
    scores?: Maybe<Array<Maybe<UserScoreStatistic>>>;
    staff?: Maybe<Array<Maybe<UserStaffStatistic>>>;
    standardDeviation: Scalars['Float']['output'];
    startYears?: Maybe<Array<Maybe<UserStartYearStatistic>>>;
    statuses?: Maybe<Array<Maybe<UserStatusStatistic>>>;
    studios?: Maybe<Array<Maybe<UserStudioStatistic>>>;
    tags?: Maybe<Array<Maybe<UserTagStatistic>>>;
    voiceActors?: Maybe<Array<Maybe<UserVoiceActorStatistic>>>;
    volumesRead: Scalars['Int']['output'];
};


export type UserStatisticsCountriesArgs = {
    limit?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<UserStatisticsSort>>>;
};


export type UserStatisticsFormatsArgs = {
    limit?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<UserStatisticsSort>>>;
};


export type UserStatisticsGenresArgs = {
    limit?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<UserStatisticsSort>>>;
};


export type UserStatisticsLengthsArgs = {
    limit?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<UserStatisticsSort>>>;
};


export type UserStatisticsReleaseYearsArgs = {
    limit?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<UserStatisticsSort>>>;
};


export type UserStatisticsScoresArgs = {
    limit?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<UserStatisticsSort>>>;
};


export type UserStatisticsStaffArgs = {
    limit?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<UserStatisticsSort>>>;
};


export type UserStatisticsStartYearsArgs = {
    limit?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<UserStatisticsSort>>>;
};


export type UserStatisticsStatusesArgs = {
    limit?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<UserStatisticsSort>>>;
};


export type UserStatisticsStudiosArgs = {
    limit?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<UserStatisticsSort>>>;
};


export type UserStatisticsTagsArgs = {
    limit?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<UserStatisticsSort>>>;
};


export type UserStatisticsVoiceActorsArgs = {
    limit?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<UserStatisticsSort>>>;
};

/** User statistics sort enum */
export type UserStatisticsSort =
    | 'COUNT'
    | 'COUNT_DESC'
    | 'ID'
    | 'ID_DESC'
    | 'MEAN_SCORE'
    | 'MEAN_SCORE_DESC'
    | 'PROGRESS'
    | 'PROGRESS_DESC';

/** A user's statistics */
export type UserStats = {
    activityHistory?: Maybe<Array<Maybe<UserActivityHistory>>>;
    animeListScores?: Maybe<ListScoreStats>;
    animeScoreDistribution?: Maybe<Array<Maybe<ScoreDistribution>>>;
    animeStatusDistribution?: Maybe<Array<Maybe<StatusDistribution>>>;
    /** The amount of manga chapters the user has read */
    chaptersRead?: Maybe<Scalars['Int']['output']>;
    favouredActors?: Maybe<Array<Maybe<StaffStats>>>;
    favouredFormats?: Maybe<Array<Maybe<FormatStats>>>;
    favouredGenres?: Maybe<Array<Maybe<GenreStats>>>;
    favouredGenresOverview?: Maybe<Array<Maybe<GenreStats>>>;
    favouredStaff?: Maybe<Array<Maybe<StaffStats>>>;
    favouredStudios?: Maybe<Array<Maybe<StudioStats>>>;
    favouredTags?: Maybe<Array<Maybe<TagStats>>>;
    favouredYears?: Maybe<Array<Maybe<YearStats>>>;
    mangaListScores?: Maybe<ListScoreStats>;
    mangaScoreDistribution?: Maybe<Array<Maybe<ScoreDistribution>>>;
    mangaStatusDistribution?: Maybe<Array<Maybe<StatusDistribution>>>;
    /** The amount of anime the user has watched in minutes */
    watchedTime?: Maybe<Scalars['Int']['output']>;
};

export type UserStatusStatistic = {
    chaptersRead: Scalars['Int']['output'];
    count: Scalars['Int']['output'];
    meanScore: Scalars['Float']['output'];
    mediaIds: Array<Maybe<Scalars['Int']['output']>>;
    minutesWatched: Scalars['Int']['output'];
    status?: Maybe<MediaListStatus>;
};

export type UserStudioStatistic = {
    chaptersRead: Scalars['Int']['output'];
    count: Scalars['Int']['output'];
    meanScore: Scalars['Float']['output'];
    mediaIds: Array<Maybe<Scalars['Int']['output']>>;
    minutesWatched: Scalars['Int']['output'];
    studio?: Maybe<Studio>;
};

export type UserTagStatistic = {
    chaptersRead: Scalars['Int']['output'];
    count: Scalars['Int']['output'];
    meanScore: Scalars['Float']['output'];
    mediaIds: Array<Maybe<Scalars['Int']['output']>>;
    minutesWatched: Scalars['Int']['output'];
    tag?: Maybe<MediaTag>;
};

/** The language the user wants to see media titles in */
export type UserTitleLanguage =
/** The official english title */
    | 'ENGLISH'
    /** The official english title, stylised by media creator */
    | 'ENGLISH_STYLISED'
    /** Official title in it's native language */
    | 'NATIVE'
    /** Official title in it's native language, stylised by media creator */
    | 'NATIVE_STYLISED'
    /** The romanization of the native language title */
    | 'ROMAJI'
    /** The romanization of the native language title, stylised by media creator */
    | 'ROMAJI_STYLISED';

export type UserVoiceActorStatistic = {
    chaptersRead: Scalars['Int']['output'];
    characterIds: Array<Maybe<Scalars['Int']['output']>>;
    count: Scalars['Int']['output'];
    meanScore: Scalars['Float']['output'];
    mediaIds: Array<Maybe<Scalars['Int']['output']>>;
    minutesWatched: Scalars['Int']['output'];
    voiceActor?: Maybe<Staff>;
};

/** User's year statistics */
export type YearStats = {
    amount?: Maybe<Scalars['Int']['output']>;
    meanScore?: Maybe<Scalars['Int']['output']>;
    year?: Maybe<Scalars['Int']['output']>;
};

export type UpdateEntryMutationVariables = Exact<{
    mediaId?: InputMaybe<Scalars['Int']['input']>;
    status?: InputMaybe<MediaListStatus>;
    score?: InputMaybe<Scalars['Float']['input']>;
    progress?: InputMaybe<Scalars['Int']['input']>;
    repeat?: InputMaybe<Scalars['Int']['input']>;
    private?: InputMaybe<Scalars['Boolean']['input']>;
    notes?: InputMaybe<Scalars['String']['input']>;
    hiddenFromStatusLists?: InputMaybe<Scalars['Boolean']['input']>;
    startedAt?: InputMaybe<FuzzyDateInput>;
    completedAt?: InputMaybe<FuzzyDateInput>;
}>;


export type UpdateEntryMutation = { SaveMediaListEntry?: { id: number } | null };

export type UpdateMediaListEntryProgressMutationVariables = Exact<{
    mediaId?: InputMaybe<Scalars["Int"]["input"]>;
    progress?: InputMaybe<Scalars["Int"]["input"]>;
}>;


export type UpdateMediaListEntryProgressMutation = { SaveMediaListEntry?: { id: number } | null };

export type DeleteEntryMutationVariables = Exact<{
    mediaListEntryId?: InputMaybe<Scalars['Int']['input']>;
}>;


export type DeleteEntryMutation = { DeleteMediaListEntry?: { deleted?: boolean | null } | null };

export type AnimeCollectionQueryVariables = Exact<{
    userName?: InputMaybe<Scalars['String']['input']>;
}>;


export type AnimeCollectionQuery = {
    MediaListCollection?: {
        lists?: Array<{
            status?: MediaListStatus | null, entries?: Array<{
                id: number,
                score?: number | null,
                progress?: number | null,
                status?: MediaListStatus | null,
                notes?: string | null,
                repeat?: number | null,
                private?: boolean | null,
                startedAt?: { year?: number | null, month?: number | null, day?: number | null } | null,
                completedAt?: { year?: number | null, month?: number | null, day?: number | null } | null,
                media?: {
                    id: number,
                    idMal?: number | null,
                    siteUrl?: string | null,
                    status?: MediaStatus | null,
                    season?: MediaSeason | null,
                    type?: MediaType | null,
                    format?: MediaFormat | null,
                    bannerImage?: string | null,
                    episodes?: number | null,
                    synonyms?: Array<string | null> | null,
                    isAdult?: boolean | null,
                    countryOfOrigin?: any | null,
                    title?: {
                        userPreferred?: string | null,
                        romaji?: string | null,
                        english?: string | null,
                        native?: string | null
                    } | null,
                    coverImage?: {
                        extraLarge?: string | null,
                        large?: string | null,
                        medium?: string | null,
                        color?: string | null
                    } | null,
                    startDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
                    endDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
                    nextAiringEpisode?: { airingAt: number, timeUntilAiring: number, episode: number } | null,
                    relations?: {
                        edges?: Array<{
                            relationType?: MediaRelation | null,
                            node?: {
                                id: number,
                                idMal?: number | null,
                                siteUrl?: string | null,
                                status?: MediaStatus | null,
                                season?: MediaSeason | null,
                                type?: MediaType | null,
                                format?: MediaFormat | null,
                                bannerImage?: string | null,
                                episodes?: number | null,
                                synonyms?: Array<string | null> | null,
                                isAdult?: boolean | null,
                                countryOfOrigin?: any | null,
                                title?: {
                                    userPreferred?: string | null,
                                    romaji?: string | null,
                                    english?: string | null,
                                    native?: string | null
                                } | null,
                                coverImage?: {
                                    extraLarge?: string | null,
                                    large?: string | null,
                                    medium?: string | null,
                                    color?: string | null
                                } | null,
                                startDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
                                endDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
                                nextAiringEpisode?: {
                                    airingAt: number,
                                    timeUntilAiring: number,
                                    episode: number
                                } | null
                            } | null
                        } | null> | null
                    } | null
                } | null
            } | null> | null
        } | null> | null
    } | null
};

export type SearchAnimeShortMediaQueryVariables = Exact<{
    page?: InputMaybe<Scalars['Int']['input']>;
    perPage?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<MediaSort>> | InputMaybe<MediaSort>>;
    search?: InputMaybe<Scalars['String']['input']>;
    status?: InputMaybe<Array<InputMaybe<MediaStatus>> | InputMaybe<MediaStatus>>;
}>;


export type SearchAnimeShortMediaQuery = {
    Page?: {
        pageInfo?: { hasNextPage?: boolean | null } | null,
        media?: Array<{
            id: number,
            idMal?: number | null,
            siteUrl?: string | null,
            status?: MediaStatus | null,
            season?: MediaSeason | null,
            type?: MediaType | null,
            format?: MediaFormat | null,
            bannerImage?: string | null,
            episodes?: number | null,
            synonyms?: Array<string | null> | null,
            isAdult?: boolean | null,
            countryOfOrigin?: any | null,
            title?: {
                userPreferred?: string | null,
                romaji?: string | null,
                english?: string | null,
                native?: string | null
            } | null,
            coverImage?: {
                extraLarge?: string | null,
                large?: string | null,
                medium?: string | null,
                color?: string | null
            } | null,
            startDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
            endDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
            nextAiringEpisode?: { airingAt: number, timeUntilAiring: number, episode: number } | null
        } | null> | null
    } | null
};

export type BasicMediaByMalIdQueryVariables = Exact<{
    id?: InputMaybe<Scalars['Int']['input']>;
}>;


export type BasicMediaByMalIdQuery = {
    Media?: {
        id: number,
        idMal?: number | null,
        siteUrl?: string | null,
        status?: MediaStatus | null,
        season?: MediaSeason | null,
        type?: MediaType | null,
        format?: MediaFormat | null,
        bannerImage?: string | null,
        episodes?: number | null,
        synonyms?: Array<string | null> | null,
        isAdult?: boolean | null,
        countryOfOrigin?: any | null,
        title?: {
            userPreferred?: string | null,
            romaji?: string | null,
            english?: string | null,
            native?: string | null
        } | null,
        coverImage?: {
            extraLarge?: string | null,
            large?: string | null,
            medium?: string | null,
            color?: string | null
        } | null,
        startDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
        endDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
        nextAiringEpisode?: { airingAt: number, timeUntilAiring: number, episode: number } | null
    } | null
};

export type BasicMediaByIdQueryVariables = Exact<{
    id?: InputMaybe<Scalars['Int']['input']>;
}>;


export type BasicMediaByIdQuery = {
    Media?: {
        id: number,
        idMal?: number | null,
        siteUrl?: string | null,
        status?: MediaStatus | null,
        season?: MediaSeason | null,
        type?: MediaType | null,
        format?: MediaFormat | null,
        bannerImage?: string | null,
        episodes?: number | null,
        synonyms?: Array<string | null> | null,
        isAdult?: boolean | null,
        countryOfOrigin?: any | null,
        title?: {
            userPreferred?: string | null,
            romaji?: string | null,
            english?: string | null,
            native?: string | null
        } | null,
        coverImage?: {
            extraLarge?: string | null,
            large?: string | null,
            medium?: string | null,
            color?: string | null
        } | null,
        startDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
        endDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
        nextAiringEpisode?: { airingAt: number, timeUntilAiring: number, episode: number } | null
    } | null
};

export type BaseMediaByIdQueryVariables = Exact<{
    id?: InputMaybe<Scalars['Int']['input']>;
}>;


export type BaseMediaByIdQuery = {
    Media?: {
        id: number,
        idMal?: number | null,
        siteUrl?: string | null,
        status?: MediaStatus | null,
        season?: MediaSeason | null,
        type?: MediaType | null,
        format?: MediaFormat | null,
        bannerImage?: string | null,
        episodes?: number | null,
        synonyms?: Array<string | null> | null,
        isAdult?: boolean | null,
        countryOfOrigin?: any | null,
        title?: {
            userPreferred?: string | null,
            romaji?: string | null,
            english?: string | null,
            native?: string | null
        } | null,
        trailer?: { id?: string | null, site?: string | null, thumbnail?: string | null } | null,
        coverImage?: {
            extraLarge?: string | null,
            large?: string | null,
            medium?: string | null,
            color?: string | null
        } | null,
        startDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
        endDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
        nextAiringEpisode?: { airingAt: number, timeUntilAiring: number, episode: number } | null,
        relations?: {
            edges?: Array<{
                relationType?: MediaRelation | null,
                node?: {
                    id: number,
                    idMal?: number | null,
                    siteUrl?: string | null,
                    status?: MediaStatus | null,
                    season?: MediaSeason | null,
                    type?: MediaType | null,
                    format?: MediaFormat | null,
                    bannerImage?: string | null,
                    episodes?: number | null,
                    synonyms?: Array<string | null> | null,
                    isAdult?: boolean | null,
                    countryOfOrigin?: any | null,
                    title?: {
                        userPreferred?: string | null,
                        romaji?: string | null,
                        english?: string | null,
                        native?: string | null
                    } | null,
                    coverImage?: {
                        extraLarge?: string | null,
                        large?: string | null,
                        medium?: string | null,
                        color?: string | null
                    } | null,
                    startDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
                    endDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
                    nextAiringEpisode?: { airingAt: number, timeUntilAiring: number, episode: number } | null
                } | null
            } | null> | null
        } | null
    } | null
};

export type MediaDetailsByIdQueryVariables = Exact<{
    id?: InputMaybe<Scalars['Int']['input']>;
}>;


export type MediaDetailsByIdQuery = {
    Media?: {
        siteUrl?: string | null,
        id: number,
        duration?: number | null,
        genres?: Array<string | null> | null,
        averageScore?: number | null,
        popularity?: number | null,
        meanScore?: number | null,
        description?: string | null,
        startDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
        endDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
        studios?: { nodes?: Array<{ name: string } | null> | null } | null,
        rankings?: Array<{
            context: string,
            type: MediaRankType,
            rank: number,
            year?: number | null,
            format: MediaFormat,
            allTime?: boolean | null,
            season?: MediaSeason | null
        } | null> | null,
        trailer?: { id?: string | null, site?: string | null, thumbnail?: string | null } | null,
        recommendations?: {
            edges?: Array<{
                node?: {
                    mediaRecommendation?: {
                        id: number,
                        bannerImage?: string | null,
                        coverImage?: {
                            extraLarge?: string | null,
                            large?: string | null,
                            medium?: string | null,
                            color?: string | null
                        } | null,
                        title?: {
                            romaji?: string | null,
                            english?: string | null,
                            native?: string | null,
                            userPreferred?: string | null
                        } | null
                    } | null
                } | null
            } | null> | null
        } | null
    } | null
};

export type CompleteMediaByIdQueryVariables = Exact<{
    id?: InputMaybe<Scalars['Int']['input']>;
}>;


export type CompleteMediaByIdQuery = {
    Media?: {
        id: number,
        idMal?: number | null,
        siteUrl?: string | null,
        status?: MediaStatus | null,
        season?: MediaSeason | null,
        type?: MediaType | null,
        format?: MediaFormat | null,
        bannerImage?: string | null,
        episodes?: number | null,
        synonyms?: Array<string | null> | null,
        isAdult?: boolean | null,
        countryOfOrigin?: any | null,
        duration?: number | null,
        genres?: Array<string | null> | null,
        averageScore?: number | null,
        popularity?: number | null,
        meanScore?: number | null,
        title?: {
            userPreferred?: string | null,
            romaji?: string | null,
            english?: string | null,
            native?: string | null
        } | null,
        coverImage?: {
            extraLarge?: string | null,
            large?: string | null,
            medium?: string | null,
            color?: string | null
        } | null,
        startDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
        endDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
        nextAiringEpisode?: { airingAt: number, timeUntilAiring: number, episode: number } | null,
        relations?: {
            edges?: Array<{
                relationType?: MediaRelation | null,
                node?: {
                    id: number,
                    idMal?: number | null,
                    siteUrl?: string | null,
                    status?: MediaStatus | null,
                    season?: MediaSeason | null,
                    type?: MediaType | null,
                    format?: MediaFormat | null,
                    bannerImage?: string | null,
                    episodes?: number | null,
                    synonyms?: Array<string | null> | null,
                    isAdult?: boolean | null,
                    countryOfOrigin?: any | null,
                    title?: {
                        userPreferred?: string | null,
                        romaji?: string | null,
                        english?: string | null,
                        native?: string | null
                    } | null,
                    coverImage?: {
                        extraLarge?: string | null,
                        large?: string | null,
                        medium?: string | null,
                        color?: string | null
                    } | null,
                    startDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
                    endDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
                    nextAiringEpisode?: { airingAt: number, timeUntilAiring: number, episode: number } | null
                } | null
            } | null> | null
        } | null
    } | null
};

export type ListMediaQueryVariables = Exact<{
    page?: InputMaybe<Scalars['Int']['input']>;
    search?: InputMaybe<Scalars['String']['input']>;
    perPage?: InputMaybe<Scalars['Int']['input']>;
    sort?: InputMaybe<Array<InputMaybe<MediaSort>> | InputMaybe<MediaSort>>;
    status?: InputMaybe<Array<InputMaybe<MediaStatus>> | InputMaybe<MediaStatus>>;
    genres?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>> | InputMaybe<Scalars['String']['input']>>;
    averageScore_greater?: InputMaybe<Scalars['Int']['input']>;
    season?: InputMaybe<MediaSeason>;
    seasonYear?: InputMaybe<Scalars['Int']['input']>;
    format?: InputMaybe<MediaFormat>;
}>;


export type ListMediaQuery = {
    Page?: {
        pageInfo?: {
            hasNextPage?: boolean | null,
            total?: number | null,
            perPage?: number | null,
            currentPage?: number | null,
            lastPage?: number | null
        } | null,
        media?: Array<{
            id: number,
            idMal?: number | null,
            siteUrl?: string | null,
            status?: MediaStatus | null,
            season?: MediaSeason | null,
            type?: MediaType | null,
            format?: MediaFormat | null,
            bannerImage?: string | null,
            episodes?: number | null,
            synonyms?: Array<string | null> | null,
            isAdult?: boolean | null,
            countryOfOrigin?: any | null,
            title?: {
                userPreferred?: string | null,
                romaji?: string | null,
                english?: string | null,
                native?: string | null
            } | null,
            coverImage?: {
                extraLarge?: string | null,
                large?: string | null,
                medium?: string | null,
                color?: string | null
            } | null,
            startDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
            endDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
            nextAiringEpisode?: { airingAt: number, timeUntilAiring: number, episode: number } | null
        } | null> | null
    } | null
};

export type ListRecentMediaQueryVariables = Exact<{
    page?: InputMaybe<Scalars['Int']['input']>;
    perPage?: InputMaybe<Scalars['Int']['input']>;
    airingAt_greater?: InputMaybe<Scalars['Int']['input']>;
    airingAt_lesser?: InputMaybe<Scalars['Int']['input']>;
}>;


export type ListRecentMediaQuery = {
    Page?: {
        pageInfo?: {
            hasNextPage?: boolean | null,
            total?: number | null,
            perPage?: number | null,
            currentPage?: number | null,
            lastPage?: number | null
        } | null,
        airingSchedules?: Array<{
            id: number,
            airingAt: number,
            episode: number,
            timeUntilAiring: number,
            media?: {
                id: number,
                idMal?: number | null,
                siteUrl?: string | null,
                status?: MediaStatus | null,
                season?: MediaSeason | null,
                type?: MediaType | null,
                format?: MediaFormat | null,
                bannerImage?: string | null,
                episodes?: number | null,
                synonyms?: Array<string | null> | null,
                isAdult?: boolean | null,
                countryOfOrigin?: any | null,
                title?: {
                    userPreferred?: string | null,
                    romaji?: string | null,
                    english?: string | null,
                    native?: string | null
                } | null,
                coverImage?: {
                    extraLarge?: string | null,
                    large?: string | null,
                    medium?: string | null,
                    color?: string | null
                } | null,
                startDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
                endDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
                nextAiringEpisode?: { airingAt: number, timeUntilAiring: number, episode: number } | null
            } | null
        } | null> | null
    } | null
};

export type BasicMediaFragment = {
    id: number,
    idMal?: number | null,
    siteUrl?: string | null,
    status?: MediaStatus | null,
    season?: MediaSeason | null,
    type?: MediaType | null,
    format?: MediaFormat | null,
    bannerImage?: string | null,
    episodes?: number | null,
    synonyms?: Array<string | null> | null,
    isAdult?: boolean | null,
    countryOfOrigin?: any | null,
    title?: {
        userPreferred?: string | null,
        romaji?: string | null,
        english?: string | null,
        native?: string | null
    } | null,
    coverImage?: {
        extraLarge?: string | null,
        large?: string | null,
        medium?: string | null,
        color?: string | null
    } | null,
    startDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
    endDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
    nextAiringEpisode?: { airingAt: number, timeUntilAiring: number, episode: number } | null
};

export type BaseMediaFragment = {
    id: number,
    idMal?: number | null,
    siteUrl?: string | null,
    status?: MediaStatus | null,
    season?: MediaSeason | null,
    type?: MediaType | null,
    format?: MediaFormat | null,
    bannerImage?: string | null,
    episodes?: number | null,
    synonyms?: Array<string | null> | null,
    isAdult?: boolean | null,
    countryOfOrigin?: any | null,
    title?: {
        userPreferred?: string | null,
        romaji?: string | null,
        english?: string | null,
        native?: string | null
    } | null,
    coverImage?: {
        extraLarge?: string | null,
        large?: string | null,
        medium?: string | null,
        color?: string | null
    } | null,
    startDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
    endDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
    nextAiringEpisode?: { airingAt: number, timeUntilAiring: number, episode: number } | null,
    relations?: {
        edges?: Array<{
            relationType?: MediaRelation | null,
            node?: {
                id: number,
                idMal?: number | null,
                siteUrl?: string | null,
                status?: MediaStatus | null,
                season?: MediaSeason | null,
                type?: MediaType | null,
                format?: MediaFormat | null,
                bannerImage?: string | null,
                episodes?: number | null,
                synonyms?: Array<string | null> | null,
                isAdult?: boolean | null,
                countryOfOrigin?: any | null,
                title?: {
                    userPreferred?: string | null,
                    romaji?: string | null,
                    english?: string | null,
                    native?: string | null
                } | null,
                coverImage?: {
                    extraLarge?: string | null,
                    large?: string | null,
                    medium?: string | null,
                    color?: string | null
                } | null,
                startDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
                endDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
                nextAiringEpisode?: { airingAt: number, timeUntilAiring: number, episode: number } | null
            } | null
        } | null> | null
    } | null
};

export type CompleteMediaFragment = {
    id: number,
    idMal?: number | null,
    siteUrl?: string | null,
    status?: MediaStatus | null,
    season?: MediaSeason | null,
    type?: MediaType | null,
    format?: MediaFormat | null,
    bannerImage?: string | null,
    episodes?: number | null,
    synonyms?: Array<string | null> | null,
    isAdult?: boolean | null,
    countryOfOrigin?: any | null,
    duration?: number | null,
    genres?: Array<string | null> | null,
    averageScore?: number | null,
    popularity?: number | null,
    meanScore?: number | null,
    title?: {
        userPreferred?: string | null,
        romaji?: string | null,
        english?: string | null,
        native?: string | null
    } | null,
    coverImage?: {
        extraLarge?: string | null,
        large?: string | null,
        medium?: string | null,
        color?: string | null
    } | null,
    startDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
    endDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
    nextAiringEpisode?: { airingAt: number, timeUntilAiring: number, episode: number } | null,
    relations?: {
        edges?: Array<{
            relationType?: MediaRelation | null,
            node?: {
                id: number,
                idMal?: number | null,
                siteUrl?: string | null,
                status?: MediaStatus | null,
                season?: MediaSeason | null,
                type?: MediaType | null,
                format?: MediaFormat | null,
                bannerImage?: string | null,
                episodes?: number | null,
                synonyms?: Array<string | null> | null,
                isAdult?: boolean | null,
                countryOfOrigin?: any | null,
                title?: {
                    userPreferred?: string | null,
                    romaji?: string | null,
                    english?: string | null,
                    native?: string | null
                } | null,
                coverImage?: {
                    extraLarge?: string | null,
                    large?: string | null,
                    medium?: string | null,
                    color?: string | null
                } | null,
                startDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
                endDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
                nextAiringEpisode?: { airingAt: number, timeUntilAiring: number, episode: number } | null
            } | null
        } | null> | null
    } | null
};

export type GetViewerQueryVariables = Exact<{ [key: string]: never; }>;


export type GetViewerQuery = {
    Viewer?: {
        name: string,
        bannerImage?: string | null,
        isBlocked?: boolean | null,
        avatar?: { large?: string | null, medium?: string | null } | null,
        options?: {
            displayAdultContent?: boolean | null,
            airingNotifications?: boolean | null,
            profileColor?: string | null
        } | null
    } | null
};

export const BasicMediaFragmentDoc = {
    "kind": "Document", "definitions": [{
        "kind": "FragmentDefinition",
        "name": { "kind": "Name", "value": "basicMedia" },
        "typeCondition": { "kind": "NamedType", "name": { "kind": "Name", "value": "Media" } },
        "selectionSet": {
            "kind": "SelectionSet",
            "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "id" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "idMal" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "siteUrl" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "status" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "version" },
                    "value": { "kind": "IntValue", "value": "2" },
                }],
            }, { "kind": "Field", "name": { "kind": "Name", "value": "season" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "type" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "format" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "bannerImage" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "episodes" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "synonyms" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "isAdult" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "countryOfOrigin" },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "title" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "userPreferred" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "romaji" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "english" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "native" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "coverImage" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "extraLarge" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "large" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "medium" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "color" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "startDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "endDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "nextAiringEpisode" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "airingAt" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "timeUntilAiring" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "episode" },
                    }],
                },
            }],
        },
    }],
} as unknown as DocumentNode<BasicMediaFragment, unknown>
export const BaseMediaFragmentDoc = {
    "kind": "Document", "definitions": [{
        "kind": "FragmentDefinition",
        "name": { "kind": "Name", "value": "baseMedia" },
        "typeCondition": { "kind": "NamedType", "name": { "kind": "Name", "value": "Media" } },
        "selectionSet": {
            "kind": "SelectionSet",
            "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "id" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "idMal" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "siteUrl" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "status" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "version" },
                    "value": { "kind": "IntValue", "value": "2" },
                }],
            }, { "kind": "Field", "name": { "kind": "Name", "value": "season" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "type" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "format" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "bannerImage" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "episodes" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "synonyms" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "isAdult" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "countryOfOrigin" },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "title" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "userPreferred" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "romaji" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "english" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "native" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "coverImage" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "extraLarge" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "large" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "medium" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "color" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "startDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "endDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "nextAiringEpisode" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "airingAt" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "timeUntilAiring" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "episode" },
                    }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "relations" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "edges" },
                        "selectionSet": {
                            "kind": "SelectionSet",
                            "selections": [{
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "relationType" },
                                "arguments": [{
                                    "kind": "Argument",
                                    "name": { "kind": "Name", "value": "version" },
                                    "value": { "kind": "IntValue", "value": "2" },
                                }],
                            }, {
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "node" },
                                "selectionSet": {
                                    "kind": "SelectionSet",
                                    "selections": [{
                                        "kind": "FragmentSpread",
                                        "name": { "kind": "Name", "value": "basicMedia" },
                                    }],
                                },
                            }],
                        },
                    }],
                },
            }],
        },
    }, {
        "kind": "FragmentDefinition",
        "name": { "kind": "Name", "value": "basicMedia" },
        "typeCondition": { "kind": "NamedType", "name": { "kind": "Name", "value": "Media" } },
        "selectionSet": {
            "kind": "SelectionSet",
            "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "id" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "idMal" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "siteUrl" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "status" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "version" },
                    "value": { "kind": "IntValue", "value": "2" },
                }],
            }, { "kind": "Field", "name": { "kind": "Name", "value": "season" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "type" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "format" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "bannerImage" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "episodes" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "synonyms" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "isAdult" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "countryOfOrigin" },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "title" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "userPreferred" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "romaji" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "english" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "native" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "coverImage" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "extraLarge" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "large" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "medium" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "color" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "startDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "endDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "nextAiringEpisode" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "airingAt" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "timeUntilAiring" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "episode" },
                    }],
                },
            }],
        },
    }],
} as unknown as DocumentNode<BaseMediaFragment, unknown>
export const CompleteMediaFragmentDoc = {
    "kind": "Document", "definitions": [{
        "kind": "FragmentDefinition",
        "name": { "kind": "Name", "value": "completeMedia" },
        "typeCondition": { "kind": "NamedType", "name": { "kind": "Name", "value": "Media" } },
        "selectionSet": {
            "kind": "SelectionSet",
            "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "id" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "idMal" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "siteUrl" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "status" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "version" },
                    "value": { "kind": "IntValue", "value": "2" },
                }],
            }, { "kind": "Field", "name": { "kind": "Name", "value": "season" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "type" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "format" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "bannerImage" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "episodes" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "synonyms" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "isAdult" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "countryOfOrigin" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "duration" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "genres" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "averageScore" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "popularity" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "meanScore" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "title" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "userPreferred" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "romaji" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "english" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "native" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "coverImage" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "extraLarge" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "large" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "medium" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "color" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "startDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "endDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "nextAiringEpisode" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "airingAt" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "timeUntilAiring" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "episode" },
                    }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "relations" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "edges" },
                        "selectionSet": {
                            "kind": "SelectionSet",
                            "selections": [{
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "relationType" },
                                "arguments": [{
                                    "kind": "Argument",
                                    "name": { "kind": "Name", "value": "version" },
                                    "value": { "kind": "IntValue", "value": "2" },
                                }],
                            }, {
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "node" },
                                "selectionSet": {
                                    "kind": "SelectionSet",
                                    "selections": [{
                                        "kind": "FragmentSpread",
                                        "name": { "kind": "Name", "value": "basicMedia" },
                                    }],
                                },
                            }],
                        },
                    }],
                },
            }],
        },
    }, {
        "kind": "FragmentDefinition",
        "name": { "kind": "Name", "value": "basicMedia" },
        "typeCondition": { "kind": "NamedType", "name": { "kind": "Name", "value": "Media" } },
        "selectionSet": {
            "kind": "SelectionSet",
            "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "id" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "idMal" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "siteUrl" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "status" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "version" },
                    "value": { "kind": "IntValue", "value": "2" },
                }],
            }, { "kind": "Field", "name": { "kind": "Name", "value": "season" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "type" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "format" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "bannerImage" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "episodes" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "synonyms" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "isAdult" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "countryOfOrigin" },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "title" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "userPreferred" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "romaji" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "english" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "native" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "coverImage" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "extraLarge" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "large" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "medium" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "color" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "startDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "endDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "nextAiringEpisode" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "airingAt" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "timeUntilAiring" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "episode" },
                    }],
                },
            }],
        },
    }],
} as unknown as DocumentNode<CompleteMediaFragment, unknown>
export const UpdateEntryDocument = {
    "kind": "Document", "definitions": [{
        "kind": "OperationDefinition",
        "operation": "mutation",
        "name": { "kind": "Name", "value": "UpdateEntry" },
        "variableDefinitions": [{
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "mediaId" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "Int" } },
        }, {
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "status" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "MediaListStatus" } },
        }, {
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "score" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "Float" } },
        }, {
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "progress" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "Int" } },
        }, {
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "repeat" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "Int" } },
        }, {
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "private" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "Boolean" } },
        }, {
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "notes" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "String" } },
        }, {
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "hiddenFromStatusLists" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "Boolean" } },
        }, {
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "startedAt" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "FuzzyDateInput" } },
        }, {
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "completedAt" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "FuzzyDateInput" } },
        }],
        "selectionSet": {
            "kind": "SelectionSet", "selections": [{
                "kind": "Field",
                "name": { "kind": "Name", "value": "SaveMediaListEntry" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "mediaId" },
                    "value": { "kind": "Variable", "name": { "kind": "Name", "value": "mediaId" } },
                }, {
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "status" },
                    "value": { "kind": "Variable", "name": { "kind": "Name", "value": "status" } },
                }, {
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "score" },
                    "value": { "kind": "Variable", "name": { "kind": "Name", "value": "score" } },
                }, {
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "progress" },
                    "value": { "kind": "Variable", "name": { "kind": "Name", "value": "progress" } },
                }, {
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "repeat" },
                    "value": { "kind": "Variable", "name": { "kind": "Name", "value": "repeat" } },
                }, {
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "private" },
                    "value": { "kind": "Variable", "name": { "kind": "Name", "value": "private" } },
                }, {
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "notes" },
                    "value": { "kind": "Variable", "name": { "kind": "Name", "value": "notes" } },
                }, {
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "hiddenFromStatusLists" },
                    "value": { "kind": "Variable", "name": { "kind": "Name", "value": "hiddenFromStatusLists" } },
                }, {
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "startedAt" },
                    "value": { "kind": "Variable", "name": { "kind": "Name", "value": "startedAt" } },
                }, {
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "completedAt" },
                    "value": { "kind": "Variable", "name": { "kind": "Name", "value": "completedAt" } },
                }],
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "id" } }],
                },
            }],
        },
    }],
} as unknown as DocumentNode<UpdateEntryMutation, UpdateEntryMutationVariables>
export const UpdateMediaListEntryProgressDocument = {
    "kind": "Document",
    "definitions": [{
        "kind": "OperationDefinition",
        "operation": "mutation",
        "name": { "kind": "Name", "value": "UpdateMediaListEntryProgress" },
        "variableDefinitions": [{
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "mediaId" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "Int" } },
        }, {
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "progress" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "Int" } },
        }],
        "selectionSet": {
            "kind": "SelectionSet",
            "selections": [{
                "kind": "Field",
                "name": { "kind": "Name", "value": "SaveMediaListEntry" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "mediaId" },
                    "value": { "kind": "Variable", "name": { "kind": "Name", "value": "mediaId" } },
                }, {
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "progress" },
                    "value": { "kind": "Variable", "name": { "kind": "Name", "value": "progress" } },
                }],
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "id" } }],
                },
            }],
        },
    }],
} as unknown as DocumentNode<UpdateMediaListEntryProgressMutation, UpdateMediaListEntryProgressMutationVariables>
export const DeleteEntryDocument = {
    "kind": "Document",
    "definitions": [{
        "kind": "OperationDefinition",
        "operation": "mutation",
        "name": { "kind": "Name", "value": "DeleteEntry" },
        "variableDefinitions": [{
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "mediaListEntryId" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "Int" } },
        }],
        "selectionSet": {
            "kind": "SelectionSet",
            "selections": [{
                "kind": "Field",
                "name": { "kind": "Name", "value": "DeleteMediaListEntry" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "id" },
                    "value": { "kind": "Variable", "name": { "kind": "Name", "value": "mediaListEntryId" } },
                }],
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "deleted" } }],
                },
            }],
        },
    }],
} as unknown as DocumentNode<DeleteEntryMutation, DeleteEntryMutationVariables>
export const AnimeCollectionDocument = {
    "kind": "Document", "definitions": [{
        "kind": "OperationDefinition",
        "operation": "query",
        "name": { "kind": "Name", "value": "AnimeCollection" },
        "variableDefinitions": [{
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "userName" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "String" } },
        }],
        "selectionSet": {
            "kind": "SelectionSet", "selections": [{
                "kind": "Field",
                "name": { "kind": "Name", "value": "MediaListCollection" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "userName" },
                    "value": { "kind": "Variable", "name": { "kind": "Name", "value": "userName" } },
                }, {
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "type" },
                    "value": { "kind": "EnumValue", "value": "ANIME" },
                }],
                "selectionSet": {
                    "kind": "SelectionSet", "selections": [{
                        "kind": "Field", "name": { "kind": "Name", "value": "lists" }, "selectionSet": {
                            "kind": "SelectionSet",
                            "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "status" } }, {
                                "kind": "Field", "name": { "kind": "Name", "value": "entries" }, "selectionSet": {
                                    "kind": "SelectionSet",
                                    "selections": [{
                                        "kind": "Field",
                                        "name": { "kind": "Name", "value": "id" },
                                    }, {
                                        "kind": "Field",
                                        "name": { "kind": "Name", "value": "score" },
                                    }, {
                                        "kind": "Field",
                                        "name": { "kind": "Name", "value": "progress" },
                                    }, {
                                        "kind": "Field",
                                        "name": { "kind": "Name", "value": "status" },
                                    }, {
                                        "kind": "Field",
                                        "name": { "kind": "Name", "value": "notes" },
                                    }, {
                                        "kind": "Field",
                                        "name": { "kind": "Name", "value": "repeat" },
                                    }, {
                                        "kind": "Field",
                                        "name": { "kind": "Name", "value": "private" },
                                    }, {
                                        "kind": "Field",
                                        "name": { "kind": "Name", "value": "startedAt" },
                                        "selectionSet": {
                                            "kind": "SelectionSet",
                                            "selections": [{
                                                "kind": "Field",
                                                "name": { "kind": "Name", "value": "year" },
                                            }, {
                                                "kind": "Field",
                                                "name": { "kind": "Name", "value": "month" },
                                            }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                                        },
                                    }, {
                                        "kind": "Field",
                                        "name": { "kind": "Name", "value": "completedAt" },
                                        "selectionSet": {
                                            "kind": "SelectionSet",
                                            "selections": [{
                                                "kind": "Field",
                                                "name": { "kind": "Name", "value": "year" },
                                            }, {
                                                "kind": "Field",
                                                "name": { "kind": "Name", "value": "month" },
                                            }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                                        },
                                    }, {
                                        "kind": "Field",
                                        "name": { "kind": "Name", "value": "media" },
                                        "selectionSet": {
                                            "kind": "SelectionSet",
                                            "selections": [{
                                                "kind": "FragmentSpread",
                                                "name": { "kind": "Name", "value": "baseMedia" },
                                            }],
                                        },
                                    }],
                                },
                            }],
                        },
                    }],
                },
            }],
        },
    }, {
        "kind": "FragmentDefinition",
        "name": { "kind": "Name", "value": "basicMedia" },
        "typeCondition": { "kind": "NamedType", "name": { "kind": "Name", "value": "Media" } },
        "selectionSet": {
            "kind": "SelectionSet",
            "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "id" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "idMal" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "siteUrl" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "status" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "version" },
                    "value": { "kind": "IntValue", "value": "2" },
                }],
            }, { "kind": "Field", "name": { "kind": "Name", "value": "season" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "type" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "format" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "bannerImage" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "episodes" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "synonyms" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "isAdult" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "countryOfOrigin" },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "title" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "userPreferred" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "romaji" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "english" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "native" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "coverImage" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "extraLarge" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "large" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "medium" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "color" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "startDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "endDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "nextAiringEpisode" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "airingAt" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "timeUntilAiring" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "episode" },
                    }],
                },
            }],
        },
    }, {
        "kind": "FragmentDefinition",
        "name": { "kind": "Name", "value": "baseMedia" },
        "typeCondition": { "kind": "NamedType", "name": { "kind": "Name", "value": "Media" } },
        "selectionSet": {
            "kind": "SelectionSet",
            "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "id" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "idMal" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "siteUrl" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "status" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "version" },
                    "value": { "kind": "IntValue", "value": "2" },
                }],
            }, { "kind": "Field", "name": { "kind": "Name", "value": "season" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "type" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "format" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "bannerImage" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "episodes" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "synonyms" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "isAdult" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "countryOfOrigin" },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "title" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "userPreferred" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "romaji" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "english" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "native" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "coverImage" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "extraLarge" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "large" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "medium" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "color" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "startDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "endDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "nextAiringEpisode" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "airingAt" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "timeUntilAiring" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "episode" },
                    }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "relations" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "edges" },
                        "selectionSet": {
                            "kind": "SelectionSet",
                            "selections": [{
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "relationType" },
                                "arguments": [{
                                    "kind": "Argument",
                                    "name": { "kind": "Name", "value": "version" },
                                    "value": { "kind": "IntValue", "value": "2" },
                                }],
                            }, {
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "node" },
                                "selectionSet": {
                                    "kind": "SelectionSet",
                                    "selections": [{
                                        "kind": "FragmentSpread",
                                        "name": { "kind": "Name", "value": "basicMedia" },
                                    }],
                                },
                            }],
                        },
                    }],
                },
            }],
        },
    }],
} as unknown as DocumentNode<AnimeCollectionQuery, AnimeCollectionQueryVariables>
export const SearchAnimeShortMediaDocument = {
    "kind": "Document", "definitions": [{
        "kind": "OperationDefinition",
        "operation": "query",
        "name": { "kind": "Name", "value": "SearchAnimeShortMedia" },
        "variableDefinitions": [{
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "page" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "Int" } },
        }, {
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "perPage" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "Int" } },
        }, {
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "sort" } },
            "type": {
                "kind": "ListType",
                "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "MediaSort" } },
            },
        }, {
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "search" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "String" } },
        }, {
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "status" } },
            "type": {
                "kind": "ListType",
                "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "MediaStatus" } },
            },
        }],
        "selectionSet": {
            "kind": "SelectionSet", "selections": [{
                "kind": "Field",
                "name": { "kind": "Name", "value": "Page" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "page" },
                    "value": { "kind": "Variable", "name": { "kind": "Name", "value": "page" } },
                }, {
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "perPage" },
                    "value": { "kind": "Variable", "name": { "kind": "Name", "value": "perPage" } },
                }],
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "pageInfo" },
                        "selectionSet": {
                            "kind": "SelectionSet",
                            "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "hasNextPage" } }],
                        },
                    }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "media" },
                        "arguments": [{
                            "kind": "Argument",
                            "name": { "kind": "Name", "value": "type" },
                            "value": { "kind": "EnumValue", "value": "ANIME" },
                        }, {
                            "kind": "Argument",
                            "name": { "kind": "Name", "value": "search" },
                            "value": { "kind": "Variable", "name": { "kind": "Name", "value": "search" } },
                        }, {
                            "kind": "Argument",
                            "name": { "kind": "Name", "value": "sort" },
                            "value": { "kind": "Variable", "name": { "kind": "Name", "value": "sort" } },
                        }, {
                            "kind": "Argument",
                            "name": { "kind": "Name", "value": "status_in" },
                            "value": { "kind": "Variable", "name": { "kind": "Name", "value": "status" } },
                        }, {
                            "kind": "Argument",
                            "name": { "kind": "Name", "value": "isAdult" },
                            "value": { "kind": "BooleanValue", "value": false },
                        }, {
                            "kind": "Argument",
                            "name": { "kind": "Name", "value": "format_not" },
                            "value": { "kind": "EnumValue", "value": "MUSIC" },
                        }],
                        "selectionSet": {
                            "kind": "SelectionSet",
                            "selections": [{
                                "kind": "FragmentSpread",
                                "name": { "kind": "Name", "value": "basicMedia" },
                            }],
                        },
                    }],
                },
            }],
        },
    }, {
        "kind": "FragmentDefinition",
        "name": { "kind": "Name", "value": "basicMedia" },
        "typeCondition": { "kind": "NamedType", "name": { "kind": "Name", "value": "Media" } },
        "selectionSet": {
            "kind": "SelectionSet",
            "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "id" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "idMal" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "siteUrl" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "status" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "version" },
                    "value": { "kind": "IntValue", "value": "2" },
                }],
            }, { "kind": "Field", "name": { "kind": "Name", "value": "season" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "type" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "format" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "bannerImage" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "episodes" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "synonyms" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "isAdult" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "countryOfOrigin" },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "title" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "userPreferred" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "romaji" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "english" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "native" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "coverImage" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "extraLarge" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "large" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "medium" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "color" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "startDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "endDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "nextAiringEpisode" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "airingAt" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "timeUntilAiring" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "episode" },
                    }],
                },
            }],
        },
    }],
} as unknown as DocumentNode<SearchAnimeShortMediaQuery, SearchAnimeShortMediaQueryVariables>
export const BasicMediaByMalIdDocument = {
    "kind": "Document",
    "definitions": [{
        "kind": "OperationDefinition",
        "operation": "query",
        "name": { "kind": "Name", "value": "BasicMediaByMalId" },
        "variableDefinitions": [{
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "id" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "Int" } },
        }],
        "selectionSet": {
            "kind": "SelectionSet",
            "selections": [{
                "kind": "Field",
                "name": { "kind": "Name", "value": "Media" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "idMal" },
                    "value": { "kind": "Variable", "name": { "kind": "Name", "value": "id" } },
                }, {
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "type" },
                    "value": { "kind": "EnumValue", "value": "ANIME" },
                }],
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "FragmentSpread", "name": { "kind": "Name", "value": "basicMedia" } }],
                },
            }],
        },
    }, {
        "kind": "FragmentDefinition",
        "name": { "kind": "Name", "value": "basicMedia" },
        "typeCondition": { "kind": "NamedType", "name": { "kind": "Name", "value": "Media" } },
        "selectionSet": {
            "kind": "SelectionSet",
            "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "id" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "idMal" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "siteUrl" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "status" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "version" },
                    "value": { "kind": "IntValue", "value": "2" },
                }],
            }, { "kind": "Field", "name": { "kind": "Name", "value": "season" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "type" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "format" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "bannerImage" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "episodes" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "synonyms" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "isAdult" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "countryOfOrigin" },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "title" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "userPreferred" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "romaji" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "english" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "native" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "coverImage" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "extraLarge" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "large" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "medium" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "color" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "startDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "endDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "nextAiringEpisode" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "airingAt" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "timeUntilAiring" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "episode" },
                    }],
                },
            }],
        },
    }],
} as unknown as DocumentNode<BasicMediaByMalIdQuery, BasicMediaByMalIdQueryVariables>
export const BasicMediaByIdDocument = {
    "kind": "Document",
    "definitions": [{
        "kind": "OperationDefinition",
        "operation": "query",
        "name": { "kind": "Name", "value": "BasicMediaById" },
        "variableDefinitions": [{
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "id" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "Int" } },
        }],
        "selectionSet": {
            "kind": "SelectionSet",
            "selections": [{
                "kind": "Field",
                "name": { "kind": "Name", "value": "Media" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "id" },
                    "value": { "kind": "Variable", "name": { "kind": "Name", "value": "id" } },
                }, {
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "type" },
                    "value": { "kind": "EnumValue", "value": "ANIME" },
                }],
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "FragmentSpread", "name": { "kind": "Name", "value": "basicMedia" } }],
                },
            }],
        },
    }, {
        "kind": "FragmentDefinition",
        "name": { "kind": "Name", "value": "basicMedia" },
        "typeCondition": { "kind": "NamedType", "name": { "kind": "Name", "value": "Media" } },
        "selectionSet": {
            "kind": "SelectionSet",
            "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "id" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "idMal" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "siteUrl" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "status" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "version" },
                    "value": { "kind": "IntValue", "value": "2" },
                }],
            }, { "kind": "Field", "name": { "kind": "Name", "value": "season" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "type" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "format" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "bannerImage" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "episodes" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "synonyms" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "isAdult" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "countryOfOrigin" },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "title" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "userPreferred" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "romaji" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "english" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "native" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "coverImage" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "extraLarge" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "large" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "medium" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "color" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "startDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "endDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "nextAiringEpisode" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "airingAt" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "timeUntilAiring" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "episode" },
                    }],
                },
            }],
        },
    }],
} as unknown as DocumentNode<BasicMediaByIdQuery, BasicMediaByIdQueryVariables>
export const BaseMediaByIdDocument = {
    "kind": "Document",
    "definitions": [{
        "kind": "OperationDefinition",
        "operation": "query",
        "name": { "kind": "Name", "value": "BaseMediaById" },
        "variableDefinitions": [{
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "id" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "Int" } },
        }],
        "selectionSet": {
            "kind": "SelectionSet",
            "selections": [{
                "kind": "Field",
                "name": { "kind": "Name", "value": "Media" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "id" },
                    "value": { "kind": "Variable", "name": { "kind": "Name", "value": "id" } },
                }, {
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "type" },
                    "value": { "kind": "EnumValue", "value": "ANIME" },
                }],
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "FragmentSpread", "name": { "kind": "Name", "value": "baseMedia" } }],
                },
            }],
        },
    }, {
        "kind": "FragmentDefinition",
        "name": { "kind": "Name", "value": "basicMedia" },
        "typeCondition": { "kind": "NamedType", "name": { "kind": "Name", "value": "Media" } },
        "selectionSet": {
            "kind": "SelectionSet",
            "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "id" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "idMal" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "siteUrl" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "status" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "version" },
                    "value": { "kind": "IntValue", "value": "2" },
                }],
            }, { "kind": "Field", "name": { "kind": "Name", "value": "season" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "type" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "format" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "bannerImage" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "episodes" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "synonyms" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "isAdult" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "countryOfOrigin" },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "title" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "userPreferred" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "romaji" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "english" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "native" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "coverImage" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "extraLarge" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "large" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "medium" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "color" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "startDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "endDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "nextAiringEpisode" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "airingAt" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "timeUntilAiring" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "episode" },
                    }],
                },
            }],
        },
    }, {
        "kind": "FragmentDefinition",
        "name": { "kind": "Name", "value": "baseMedia" },
        "typeCondition": { "kind": "NamedType", "name": { "kind": "Name", "value": "Media" } },
        "selectionSet": {
            "kind": "SelectionSet",
            "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "id" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "idMal" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "siteUrl" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "status" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "version" },
                    "value": { "kind": "IntValue", "value": "2" },
                }],
            }, { "kind": "Field", "name": { "kind": "Name", "value": "season" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "type" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "format" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "bannerImage" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "episodes" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "synonyms" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "isAdult" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "countryOfOrigin" },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "title" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "userPreferred" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "romaji" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "english" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "native" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "coverImage" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "extraLarge" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "large" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "medium" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "color" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "startDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "endDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "nextAiringEpisode" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "airingAt" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "timeUntilAiring" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "episode" },
                    }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "relations" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "edges" },
                        "selectionSet": {
                            "kind": "SelectionSet",
                            "selections": [{
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "relationType" },
                                "arguments": [{
                                    "kind": "Argument",
                                    "name": { "kind": "Name", "value": "version" },
                                    "value": { "kind": "IntValue", "value": "2" },
                                }],
                            }, {
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "node" },
                                "selectionSet": {
                                    "kind": "SelectionSet",
                                    "selections": [{
                                        "kind": "FragmentSpread",
                                        "name": { "kind": "Name", "value": "basicMedia" },
                                    }],
                                },
                            }],
                        },
                    }],
                },
            }],
        },
    }],
} as unknown as DocumentNode<BaseMediaByIdQuery, BaseMediaByIdQueryVariables>
export const MediaDetailsByIdDocument = {
    "kind": "Document", "definitions": [{
        "kind": "OperationDefinition",
        "operation": "query",
        "name": { "kind": "Name", "value": "MediaDetailsById" },
        "variableDefinitions": [{
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "id" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "Int" } },
        }],
        "selectionSet": {
            "kind": "SelectionSet", "selections": [{
                "kind": "Field",
                "name": { "kind": "Name", "value": "Media" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "id" },
                    "value": { "kind": "Variable", "name": { "kind": "Name", "value": "id" } },
                }, {
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "type" },
                    "value": { "kind": "EnumValue", "value": "ANIME" },
                }],
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "siteUrl" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "id" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "duration" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "genres" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "averageScore" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "popularity" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "meanScore" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "description" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "startDate" },
                        "selectionSet": {
                            "kind": "SelectionSet",
                            "selections": [{
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "year" },
                            }, { "kind": "Field", "name": { "kind": "Name", "value": "month" } }, {
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "day" },
                            }],
                        },
                    }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "endDate" },
                        "selectionSet": {
                            "kind": "SelectionSet",
                            "selections": [{
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "year" },
                            }, { "kind": "Field", "name": { "kind": "Name", "value": "month" } }, {
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "day" },
                            }],
                        },
                    }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "studios" },
                        "arguments": [{
                            "kind": "Argument",
                            "name": { "kind": "Name", "value": "isMain" },
                            "value": { "kind": "BooleanValue", "value": true },
                        }],
                        "selectionSet": {
                            "kind": "SelectionSet",
                            "selections": [{
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "nodes" },
                                "selectionSet": {
                                    "kind": "SelectionSet",
                                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "name" } }],
                                },
                            }],
                        },
                    }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "rankings" },
                        "selectionSet": {
                            "kind": "SelectionSet",
                            "selections": [{
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "context" },
                            }, { "kind": "Field", "name": { "kind": "Name", "value": "type" } }, {
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "rank" },
                            }, { "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "format" },
                            }, { "kind": "Field", "name": { "kind": "Name", "value": "allTime" } }, {
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "season" },
                            }],
                        },
                    }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "recommendations" },
                        "arguments": [{
                            "kind": "Argument",
                            "name": { "kind": "Name", "value": "page" },
                            "value": { "kind": "IntValue", "value": "1" },
                        }, {
                            "kind": "Argument",
                            "name": { "kind": "Name", "value": "perPage" },
                            "value": { "kind": "IntValue", "value": "8" },
                        }, {
                            "kind": "Argument",
                            "name": { "kind": "Name", "value": "sort" },
                            "value": { "kind": "EnumValue", "value": "RATING_DESC" },
                        }],
                        "selectionSet": {
                            "kind": "SelectionSet", "selections": [{
                                "kind": "Field", "name": { "kind": "Name", "value": "edges" }, "selectionSet": {
                                    "kind": "SelectionSet", "selections": [{
                                        "kind": "Field",
                                        "name": { "kind": "Name", "value": "node" },
                                        "selectionSet": {
                                            "kind": "SelectionSet",
                                            "selections": [{
                                                "kind": "Field",
                                                "name": { "kind": "Name", "value": "mediaRecommendation" },
                                                "selectionSet": {
                                                    "kind": "SelectionSet",
                                                    "selections": [{
                                                        "kind": "Field",
                                                        "name": { "kind": "Name", "value": "id" },
                                                    }, {
                                                        "kind": "Field",
                                                        "name": { "kind": "Name", "value": "coverImage" },
                                                        "selectionSet": {
                                                            "kind": "SelectionSet",
                                                            "selections": [{
                                                                "kind": "Field",
                                                                "name": { "kind": "Name", "value": "extraLarge" },
                                                            }, {
                                                                "kind": "Field",
                                                                "name": { "kind": "Name", "value": "large" },
                                                            }, {
                                                                "kind": "Field",
                                                                "name": { "kind": "Name", "value": "medium" },
                                                            }, {
                                                                "kind": "Field",
                                                                "name": { "kind": "Name", "value": "color" },
                                                            }],
                                                        },
                                                    }, {
                                                        "kind": "Field",
                                                        "name": { "kind": "Name", "value": "bannerImage" },
                                                    }, {
                                                        "kind": "Field",
                                                        "name": { "kind": "Name", "value": "title" },
                                                        "selectionSet": {
                                                            "kind": "SelectionSet",
                                                            "selections": [{
                                                                "kind": "Field",
                                                                "name": { "kind": "Name", "value": "romaji" },
                                                            }, {
                                                                "kind": "Field",
                                                                "name": { "kind": "Name", "value": "english" },
                                                            }, {
                                                                "kind": "Field",
                                                                "name": { "kind": "Name", "value": "native" },
                                                            }, {
                                                                "kind": "Field",
                                                                "name": { "kind": "Name", "value": "userPreferred" },
                                                            }],
                                                        },
                                                    }],
                                                },
                                            }],
                                        },
                                    }],
                                },
                            }],
                        },
                    }],
                },
            }],
        },
    }],
} as unknown as DocumentNode<MediaDetailsByIdQuery, MediaDetailsByIdQueryVariables>
export const CompleteMediaByIdDocument = {
    "kind": "Document",
    "definitions": [{
        "kind": "OperationDefinition",
        "operation": "query",
        "name": { "kind": "Name", "value": "CompleteMediaById" },
        "variableDefinitions": [{
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "id" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "Int" } },
        }],
        "selectionSet": {
            "kind": "SelectionSet",
            "selections": [{
                "kind": "Field",
                "name": { "kind": "Name", "value": "Media" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "id" },
                    "value": { "kind": "Variable", "name": { "kind": "Name", "value": "id" } },
                }, {
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "type" },
                    "value": { "kind": "EnumValue", "value": "ANIME" },
                }],
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "FragmentSpread", "name": { "kind": "Name", "value": "completeMedia" } }],
                },
            }],
        },
    }, {
        "kind": "FragmentDefinition",
        "name": { "kind": "Name", "value": "basicMedia" },
        "typeCondition": { "kind": "NamedType", "name": { "kind": "Name", "value": "Media" } },
        "selectionSet": {
            "kind": "SelectionSet",
            "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "id" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "idMal" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "siteUrl" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "status" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "version" },
                    "value": { "kind": "IntValue", "value": "2" },
                }],
            }, { "kind": "Field", "name": { "kind": "Name", "value": "season" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "type" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "format" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "bannerImage" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "episodes" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "synonyms" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "isAdult" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "countryOfOrigin" },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "title" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "userPreferred" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "romaji" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "english" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "native" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "coverImage" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "extraLarge" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "large" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "medium" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "color" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "startDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "endDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "nextAiringEpisode" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "airingAt" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "timeUntilAiring" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "episode" },
                    }],
                },
            }],
        },
    }, {
        "kind": "FragmentDefinition",
        "name": { "kind": "Name", "value": "completeMedia" },
        "typeCondition": { "kind": "NamedType", "name": { "kind": "Name", "value": "Media" } },
        "selectionSet": {
            "kind": "SelectionSet",
            "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "id" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "idMal" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "siteUrl" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "status" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "version" },
                    "value": { "kind": "IntValue", "value": "2" },
                }],
            }, { "kind": "Field", "name": { "kind": "Name", "value": "season" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "type" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "format" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "bannerImage" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "episodes" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "synonyms" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "isAdult" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "countryOfOrigin" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "duration" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "genres" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "averageScore" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "popularity" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "meanScore" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "title" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "userPreferred" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "romaji" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "english" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "native" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "coverImage" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "extraLarge" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "large" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "medium" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "color" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "startDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "endDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "nextAiringEpisode" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "airingAt" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "timeUntilAiring" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "episode" },
                    }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "relations" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "edges" },
                        "selectionSet": {
                            "kind": "SelectionSet",
                            "selections": [{
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "relationType" },
                                "arguments": [{
                                    "kind": "Argument",
                                    "name": { "kind": "Name", "value": "version" },
                                    "value": { "kind": "IntValue", "value": "2" },
                                }],
                            }, {
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "node" },
                                "selectionSet": {
                                    "kind": "SelectionSet",
                                    "selections": [{
                                        "kind": "FragmentSpread",
                                        "name": { "kind": "Name", "value": "basicMedia" },
                                    }],
                                },
                            }],
                        },
                    }],
                },
            }],
        },
    }],
} as unknown as DocumentNode<CompleteMediaByIdQuery, CompleteMediaByIdQueryVariables>
export const ListMediaDocument = {
    "kind": "Document", "definitions": [{
        "kind": "OperationDefinition",
        "operation": "query",
        "name": { "kind": "Name", "value": "ListMedia" },
        "variableDefinitions": [{
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "page" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "Int" } },
        }, {
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "search" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "String" } },
        }, {
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "perPage" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "Int" } },
        }, {
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "sort" } },
            "type": {
                "kind": "ListType",
                "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "MediaSort" } },
            },
        }, {
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "status" } },
            "type": {
                "kind": "ListType",
                "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "MediaStatus" } },
            },
        }, {
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "genres" } },
            "type": {
                "kind": "ListType",
                "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "String" } },
            },
        }, {
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "averageScore_greater" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "Int" } },
        }, {
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "season" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "MediaSeason" } },
        }, {
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "seasonYear" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "Int" } },
        }, {
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "format" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "MediaFormat" } },
        }],
        "selectionSet": {
            "kind": "SelectionSet", "selections": [{
                "kind": "Field",
                "name": { "kind": "Name", "value": "Page" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "page" },
                    "value": { "kind": "Variable", "name": { "kind": "Name", "value": "page" } },
                }, {
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "perPage" },
                    "value": { "kind": "Variable", "name": { "kind": "Name", "value": "perPage" } },
                }],
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "pageInfo" },
                        "selectionSet": {
                            "kind": "SelectionSet",
                            "selections": [{
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "hasNextPage" },
                            }, { "kind": "Field", "name": { "kind": "Name", "value": "total" } }, {
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "perPage" },
                            }, {
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "currentPage" },
                            }, { "kind": "Field", "name": { "kind": "Name", "value": "lastPage" } }],
                        },
                    }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "media" },
                        "arguments": [{
                            "kind": "Argument",
                            "name": { "kind": "Name", "value": "type" },
                            "value": { "kind": "EnumValue", "value": "ANIME" },
                        }, {
                            "kind": "Argument",
                            "name": { "kind": "Name", "value": "search" },
                            "value": { "kind": "Variable", "name": { "kind": "Name", "value": "search" } },
                        }, {
                            "kind": "Argument",
                            "name": { "kind": "Name", "value": "sort" },
                            "value": { "kind": "Variable", "name": { "kind": "Name", "value": "sort" } },
                        }, {
                            "kind": "Argument",
                            "name": { "kind": "Name", "value": "status_in" },
                            "value": { "kind": "Variable", "name": { "kind": "Name", "value": "status" } },
                        }, {
                            "kind": "Argument",
                            "name": { "kind": "Name", "value": "isAdult" },
                            "value": { "kind": "BooleanValue", "value": false },
                        }, {
                            "kind": "Argument",
                            "name": { "kind": "Name", "value": "format" },
                            "value": { "kind": "Variable", "name": { "kind": "Name", "value": "format" } },
                        }, {
                            "kind": "Argument",
                            "name": { "kind": "Name", "value": "genre_in" },
                            "value": { "kind": "Variable", "name": { "kind": "Name", "value": "genres" } },
                        }, {
                            "kind": "Argument",
                            "name": { "kind": "Name", "value": "averageScore_greater" },
                            "value": { "kind": "Variable", "name": { "kind": "Name", "value": "averageScore_greater" } },
                        }, {
                            "kind": "Argument",
                            "name": { "kind": "Name", "value": "season" },
                            "value": { "kind": "Variable", "name": { "kind": "Name", "value": "season" } },
                        }, {
                            "kind": "Argument",
                            "name": { "kind": "Name", "value": "seasonYear" },
                            "value": { "kind": "Variable", "name": { "kind": "Name", "value": "seasonYear" } },
                        }, {
                            "kind": "Argument",
                            "name": { "kind": "Name", "value": "format_not" },
                            "value": { "kind": "EnumValue", "value": "MUSIC" },
                        }],
                        "selectionSet": {
                            "kind": "SelectionSet",
                            "selections": [{
                                "kind": "FragmentSpread",
                                "name": { "kind": "Name", "value": "basicMedia" },
                            }],
                        },
                    }],
                },
            }],
        },
    }, {
        "kind": "FragmentDefinition",
        "name": { "kind": "Name", "value": "basicMedia" },
        "typeCondition": { "kind": "NamedType", "name": { "kind": "Name", "value": "Media" } },
        "selectionSet": {
            "kind": "SelectionSet",
            "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "id" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "idMal" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "siteUrl" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "status" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "version" },
                    "value": { "kind": "IntValue", "value": "2" },
                }],
            }, { "kind": "Field", "name": { "kind": "Name", "value": "season" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "type" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "format" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "bannerImage" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "episodes" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "synonyms" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "isAdult" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "countryOfOrigin" },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "title" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "userPreferred" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "romaji" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "english" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "native" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "coverImage" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "extraLarge" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "large" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "medium" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "color" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "startDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "endDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "nextAiringEpisode" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "airingAt" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "timeUntilAiring" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "episode" },
                    }],
                },
            }],
        },
    }],
} as unknown as DocumentNode<ListMediaQuery, ListMediaQueryVariables>
export const ListRecentMediaDocument = {
    "kind": "Document", "definitions": [{
        "kind": "OperationDefinition",
        "operation": "query",
        "name": { "kind": "Name", "value": "ListRecentMedia" },
        "variableDefinitions": [{
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "page" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "Int" } },
        }, {
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "perPage" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "Int" } },
        }, {
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "airingAt_greater" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "Int" } },
        }, {
            "kind": "VariableDefinition",
            "variable": { "kind": "Variable", "name": { "kind": "Name", "value": "airingAt_lesser" } },
            "type": { "kind": "NamedType", "name": { "kind": "Name", "value": "Int" } },
        }],
        "selectionSet": {
            "kind": "SelectionSet", "selections": [{
                "kind": "Field",
                "name": { "kind": "Name", "value": "Page" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "page" },
                    "value": { "kind": "Variable", "name": { "kind": "Name", "value": "page" } },
                }, {
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "perPage" },
                    "value": { "kind": "Variable", "name": { "kind": "Name", "value": "perPage" } },
                }],
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "pageInfo" },
                        "selectionSet": {
                            "kind": "SelectionSet",
                            "selections": [{
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "hasNextPage" },
                            }, { "kind": "Field", "name": { "kind": "Name", "value": "total" } }, {
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "perPage" },
                            }, {
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "currentPage" },
                            }, { "kind": "Field", "name": { "kind": "Name", "value": "lastPage" } }],
                        },
                    }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "airingSchedules" },
                        "arguments": [{
                            "kind": "Argument",
                            "name": { "kind": "Name", "value": "notYetAired" },
                            "value": { "kind": "BooleanValue", "value": false },
                        }, {
                            "kind": "Argument",
                            "name": { "kind": "Name", "value": "sort" },
                            "value": { "kind": "EnumValue", "value": "TIME_DESC" },
                        }, {
                            "kind": "Argument",
                            "name": { "kind": "Name", "value": "airingAt_greater" },
                            "value": { "kind": "Variable", "name": { "kind": "Name", "value": "airingAt_greater" } },
                        }, {
                            "kind": "Argument",
                            "name": { "kind": "Name", "value": "airingAt_lesser" },
                            "value": { "kind": "Variable", "name": { "kind": "Name", "value": "airingAt_lesser" } },
                        }],
                        "selectionSet": {
                            "kind": "SelectionSet",
                            "selections": [{
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "id" },
                            }, { "kind": "Field", "name": { "kind": "Name", "value": "airingAt" } }, {
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "episode" },
                            }, {
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "timeUntilAiring" },
                            }, {
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "media" },
                                "selectionSet": {
                                    "kind": "SelectionSet",
                                    "selections": [{
                                        "kind": "FragmentSpread",
                                        "name": { "kind": "Name", "value": "basicMedia" },
                                    }],
                                },
                            }],
                        },
                    }],
                },
            }],
        },
    }, {
        "kind": "FragmentDefinition",
        "name": { "kind": "Name", "value": "basicMedia" },
        "typeCondition": { "kind": "NamedType", "name": { "kind": "Name", "value": "Media" } },
        "selectionSet": {
            "kind": "SelectionSet",
            "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "id" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "idMal" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "siteUrl" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "status" },
                "arguments": [{
                    "kind": "Argument",
                    "name": { "kind": "Name", "value": "version" },
                    "value": { "kind": "IntValue", "value": "2" },
                }],
            }, { "kind": "Field", "name": { "kind": "Name", "value": "season" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "type" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "format" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "bannerImage" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "episodes" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "synonyms" },
            }, { "kind": "Field", "name": { "kind": "Name", "value": "isAdult" } }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "countryOfOrigin" },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "title" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "userPreferred" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "romaji" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "english" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "native" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "coverImage" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "extraLarge" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "large" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "medium" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "color" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "startDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "endDate" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "year" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "month" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "day" } }],
                },
            }, {
                "kind": "Field",
                "name": { "kind": "Name", "value": "nextAiringEpisode" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "airingAt" },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "timeUntilAiring" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "episode" },
                    }],
                },
            }],
        },
    }],
} as unknown as DocumentNode<ListRecentMediaQuery, ListRecentMediaQueryVariables>
export const GetViewerDocument = {
    "kind": "Document",
    "definitions": [{
        "kind": "OperationDefinition",
        "operation": "query",
        "name": { "kind": "Name", "value": "GetViewer" },
        "selectionSet": {
            "kind": "SelectionSet",
            "selections": [{
                "kind": "Field",
                "name": { "kind": "Name", "value": "Viewer" },
                "selectionSet": {
                    "kind": "SelectionSet",
                    "selections": [{ "kind": "Field", "name": { "kind": "Name", "value": "name" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "avatar" },
                        "selectionSet": {
                            "kind": "SelectionSet",
                            "selections": [{
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "large" },
                            }, { "kind": "Field", "name": { "kind": "Name", "value": "medium" } }],
                        },
                    }, { "kind": "Field", "name": { "kind": "Name", "value": "bannerImage" } }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "isBlocked" },
                    }, {
                        "kind": "Field",
                        "name": { "kind": "Name", "value": "options" },
                        "selectionSet": {
                            "kind": "SelectionSet",
                            "selections": [{
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "displayAdultContent" },
                            }, {
                                "kind": "Field",
                                "name": { "kind": "Name", "value": "airingNotifications" },
                            }, { "kind": "Field", "name": { "kind": "Name", "value": "profileColor" } }],
                        },
                    }],
                },
            }],
        },
    }],
} as unknown as DocumentNode<GetViewerQuery, GetViewerQueryVariables>
