import * as jtd from "jtd";

export default {
  definitions: {
    Command: {
      discriminator: "type",
      mapping: {
        JeopardyChooseQuestion: {
          properties: {
            data: {
              ref: "CommandJeopardyChooseQuestion",
            },
          },
        },
        JoinGame: {
          properties: {
            data: {
              ref: "CommandJoinGame",
            },
          },
        },
      },
    },
    CommandJeopardyChooseQuestion: {
      additionalProperties: false,
      metadata: {
        description:
          "CommandJeopardyChooseQuestion is emitted when a player chooses a question\nto answer. The server must do validation to ensure that the player is\nallowed to choose the question.\n",
      },
      optionalProperties: {},
      properties: {
        category: {
          type: "string",
        },
        question: {
          type: "string",
        },
      },
    },
    CommandJoinGame: {
      additionalProperties: false,
      metadata: {
        description:
          "CommandJoinGame is sent by a client to join a game. The client (or the\nuser) supplies a game ID and a player name. The server will respond with\nan EventJoinedGame.\n",
      },
      optionalProperties: {},
      properties: {
        gameID: {
          metadata: {
            description: "gameID is the ID of the game to join.",
          },
          type: "string",
        },
        playerName: {
          metadata: {
            description: "playerName is the wanted name of the user.",
          },
          ref: "PlayerName",
        },
      },
    },
    Error: {
      additionalProperties: false,
      metadata: {
        description: "Error is returned on every API error.\n",
      },
      optionalProperties: {},
      properties: {
        message: {
          metadata: {
            description: "Message is the error message",
          },
          type: "string",
        },
      },
    },
    Event: {
      discriminator: "type",
      mapping: {
        GameEnded: {
          properties: {
            data: {
              ref: "EventGameEnded",
            },
          },
        },
        JeopardyBeginQuestion: {
          properties: {
            data: {
              ref: "EventJeopardyBeginQuestion",
            },
          },
        },
        JeopardyTurnEnded: {
          properties: {
            data: {
              ref: "EventJeopardyTurnEnded",
            },
          },
        },
        JoinedGame: {
          properties: {
            data: {
              ref: "EventJoinedGame",
            },
          },
        },
        PlayerJoined: {
          properties: {
            data: {
              ref: "EventPlayerJoined",
            },
          },
        },
      },
    },
    EventGameEnded: {
      additionalProperties: false,
      metadata: {
        description: "EventGameEnded is emitted when the current game ends.\n",
      },
      optionalProperties: {},
      properties: {
        leaderboard: {
          ref: "Leaderboard",
        },
      },
    },
    EventJeopardyBeginQuestion: {
      additionalProperties: false,
      metadata: {
        description:
          "EventJeopardyBeginQuestion is emitted when a question begins within this\nJeopardy game. It is usually emitted once the chooser player has chosen a\ncategory and value.\n\nEach category name and question value will map to a category and question\nwithin the game data. Note that a question may repeat across multiple\ncategories.\n",
      },
      optionalProperties: {},
      properties: {
        category: {
          type: "string",
        },
        question: {
          type: "string",
        },
      },
    },
    EventJeopardyTurnEnded: {
      additionalProperties: false,
      metadata: {
        description:
          "EventJeopardyTurnEnded is emitted when a turn ends or when the game first\nstarts.\n",
      },
      optionalProperties: {},
      properties: {
        currentScore: {
          type: "float64",
        },
        isChooser: {
          type: "boolean",
        },
        leaderboard: {
          ref: "Leaderboard",
        },
      },
    },
    EventJoinedGame: {
      metadata: {
        description:
          "EventJoinedGame is emitted when the current player joins a game. It is a\nreply to CommandJoinGame and is only for the current player. Not to be\nconfused with EventPlayerJoinedGame, which is emitted when any player\njoins the current game.\n",
      },
      ref: "Game",
    },
    EventPlayerJoined: {
      additionalProperties: false,
      metadata: {
        description:
          "EventPlayerJoined is emitted when a player joins the current game.\n",
      },
      optionalProperties: {},
      properties: {
        playerName: {
          ref: "PlayerName",
        },
      },
    },
    Game: {
      discriminator: "game",
      mapping: {
        jeopardy: {
          properties: {
            data: {
              ref: "Jeopardy",
            },
          },
        },
        kahoot: {
          properties: {
            data: {
              ref: "Kahoot",
            },
          },
        },
      },
      metadata: {
        description:
          "Game is the main game object. It contains all the information about the\ngame.\n",
      },
    },
    Jeopardy: {
      additionalProperties: false,
      metadata: {
        description: "JeopardyGame is the game data for a Jeopardy game.\n",
      },
      optionalProperties: {
        moderators: {
          metadata: {
            description: "moderators enables moderators being able to join.\n",
          },
          type: "boolean",
        },
        require_name: {
          metadata: {
            description:
              "require_name, if true, will require members to input a name before\nwe can participate.\n",
          },
          type: "boolean",
        },
        score_multiplier: {
          metadata: {
            description:
              "score_multiplier is the score multiplier for each question. The\ndefault is 100.\n",
          },
          type: "float64",
        },
      },
      properties: {
        categories: {
          elements: {
            ref: "JeopardyCategory",
          },
        },
      },
    },
    JeopardyCategory: {
      additionalProperties: false,
      metadata: {
        description: "JeopardyCategory is a category in a Jeopardy game.\n",
      },
      optionalProperties: {},
      properties: {
        name: {
          metadata: {
            description: "name is the name of the category.\n",
          },
          type: "string",
        },
        questions: {
          elements: {
            ref: "JeopardyQuestion",
          },
          metadata: {
            description: "questions are the questions in the category.\n",
          },
        },
      },
    },
    JeopardyQuestion: {
      additionalProperties: false,
      metadata: {
        description: "JeopardyQuestion is a question in a Jeopardy game.\n",
      },
      optionalProperties: {},
      properties: {
        answers: {
          elements: {
            type: "string",
          },
          metadata: {
            description: "answers are the possible answers.\n",
          },
        },
        correct_answer: {
          metadata: {
            description:
              "correct_answer is the correct answer within the list of answers\nabove. The index starts at 1.\n",
          },
          type: "int32",
        },
        question: {
          metadata: {
            description: "question is the question.\n",
          },
          type: "string",
        },
      },
    },
    Kahoot: {
      additionalProperties: false,
      metadata: {
        description: "KahootGame is the game data for a Kahoot game.\n",
      },
      optionalProperties: {},
      properties: {
        questions: {
          elements: {
            ref: "KahootQuestion",
          },
          metadata: {
            description: "questions are the questions in the game.\n",
          },
        },
        time_limit: {
          metadata: {
            description:
              "time_limit is the time limit for each question. The format is in\nGo's time.Duration, e.g. 10s for 10 seconds.\n",
          },
          type: "string",
        },
      },
    },
    KahootQuestion: {
      additionalProperties: false,
      metadata: {
        description: "KahootQuestion is a question in a Kahoot game.\n",
      },
      optionalProperties: {},
      properties: {
        answers: {
          elements: {
            type: "string",
          },
          metadata: {
            description: "answers are the possible answers.\n",
          },
        },
        question: {
          metadata: {
            description: "question is the question.\n",
          },
          type: "string",
        },
      },
    },
    Leaderboard: {
      elements: {
        ref: "LeaderboardEntry",
      },
      metadata: {
        description: "Leaderboard is a list of players and their scores.\n",
      },
    },
    LeaderboardEntry: {
      additionalProperties: false,
      optionalProperties: {},
      properties: {
        playerName: {
          type: "string",
        },
        score: {
          type: "int32",
        },
      },
    },
    PlayerName: {
      metadata: {
        description: "PlayerName is the name of a player.\n",
      },
      type: "string",
    },
  },
} as jtd.Schema;
