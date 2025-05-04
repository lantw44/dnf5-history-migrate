package migrate

import (
	"context"
	"database/sql"
)

type dnf4TransItemAction int32

const (
	dnf4TransItemActionINSTALL       dnf4TransItemAction = 1  // a new package that was installed on the system
	dnf4TransItemActionDOWNGRADE     dnf4TransItemAction = 2  // an older package version that replaced previously installed version
	dnf4TransItemActionDOWNGRADED    dnf4TransItemAction = 3  // an original package version that was replaced
	dnf4TransItemActionOBSOLETE      dnf4TransItemAction = 4  //
	dnf4TransItemActionOBSOLETED     dnf4TransItemAction = 5  //
	dnf4TransItemActionUPGRADE       dnf4TransItemAction = 6  //
	dnf4TransItemActionUPGRADED      dnf4TransItemAction = 7  //
	dnf4TransItemActionREMOVE        dnf4TransItemAction = 8  // a package that was removed from the system
	dnf4TransItemActionREINSTALL     dnf4TransItemAction = 9  // a package that was reinstalled with the identical version
	dnf4TransItemActionREINSTALLED   dnf4TransItemAction = 10 // a package that was reinstalled with the identical version (old repo, for example)
	dnf4TransItemActionREASON_CHANGE dnf4TransItemAction = 11 // a package was kept on the system but it's reason has changed
)

type dnf4TransItemReason int32

const (
	dnf4TransItemReasonUNKNOWN         dnf4TransItemReason = 0
	dnf4TransItemReasonDEPENDENCY      dnf4TransItemReason = 1
	dnf4TransItemReasonUSER            dnf4TransItemReason = 2
	dnf4TransItemReasonCLEAN           dnf4TransItemReason = 3 // hawkey compatibility
	dnf4TransItemReasonWEAK_DEPENDENCY dnf4TransItemReason = 4
	dnf4TransItemReasonGROUP           dnf4TransItemReason = 5
)

type dnf4TransItemState int32

const (
	dnf4TransItemStateUNKNOWN dnf4TransItemState = 0 // default state, must be changed before save
	dnf4TransItemStateDONE    dnf4TransItemState = 1
	dnf4TransItemStateERROR   dnf4TransItemState = 2
)

type dnf4TransItem struct {
	// CREATE TABLE trans_item (
	id       int64               // INTEGER PRIMARY KEY AUTOINCREMENT,
	trans_id sql.Null[int64]     // INTEGER REFERENCES trans(id),
	item_id  sql.Null[int64]     // INTEGER REFERENCES item(id),
	repo_id  sql.Null[int64]     // INTEGER REFERENCES repo(id),
	action   dnf4TransItemAction // INTEGER NOT NULL,              /* (enum) */
	reason   dnf4TransItemReason // INTEGER NOT NULL,              /* (enum) */
	state    dnf4TransItemState  // INTEGER NOT NULL               /* (enum) */
	// );
}

type dnf5TransItemAction int32

const (
	dnf5TransItemActionInstall      dnf5TransItemAction = 1
	dnf5TransItemActionUpgrade      dnf5TransItemAction = 2
	dnf5TransItemActionDowngrade    dnf5TransItemAction = 3
	dnf5TransItemActionReinstall    dnf5TransItemAction = 4
	dnf5TransItemActionRemove       dnf5TransItemAction = 5
	dnf5TransItemActionReplaced     dnf5TransItemAction = 6
	dnf5TransItemActionReasonChange dnf5TransItemAction = 7
)

type dnf5TransItemReason int32

const (
	dnf5TransItemReasonNone           dnf5TransItemReason = 0
	dnf5TransItemReasonDependency     dnf5TransItemReason = 1
	dnf5TransItemReasonUser           dnf5TransItemReason = 2
	dnf5TransItemReasonClean          dnf5TransItemReason = 3
	dnf5TransItemReasonWeakDependency dnf5TransItemReason = 4
	dnf5TransItemReasonGroup          dnf5TransItemReason = 5
	dnf5TransItemReasonExternalUser   dnf5TransItemReason = 6
)

type dnf5TransItemState int32

const (
	dnf5TransItemStateStarted dnf5TransItemState = 1
	dnf5TransItemStateOk      dnf5TransItemState = 2
	dnf5TransItemStateError   dnf5TransItemState = 3
)

type dnf5TransItem struct {
	// CREATE TABLE "trans_item" (
	id        int64               // INTEGER,
	trans_id  sql.Null[int64]     // INTEGER,
	item_id   sql.Null[int64]     // INTEGER,
	repo_id   sql.Null[int64]     // INTEGER,
	action_id dnf5TransItemAction // INTEGER NOT NULL,             /* (enum) */
	reason_id dnf5TransItemReason // INTEGER NOT NULL,             /* (enum) */
	state_id  dnf5TransItemState  // INTEGER NOT NULL,             /* (enum) */
	//    PRIMARY KEY("id" AUTOINCREMENT),
	//    FOREIGN KEY("trans_id") REFERENCES "trans"("id"),
	//    FOREIGN KEY("item_id") REFERENCES "item"("id"),
	//    FOREIGN KEY("repo_id") REFERENCES "repo"("id"),
	//    FOREIGN KEY("action_id") REFERENCES "trans_item_action"("id"),
	//    FOREIGN KEY("reason_id") REFERENCES "trans_item_reason"("id"),
	//    FOREIGN KEY("state_id") REFERENCES "trans_item_state"("id")
	//);
}

func castTransItemAction(a dnf4TransItemAction) (dnf5TransItemAction, error) {
	switch a {
	case dnf4TransItemActionINSTALL:
		return dnf5TransItemActionInstall, nil
	case dnf4TransItemActionDOWNGRADE:
		return dnf5TransItemActionDowngrade, nil
	case dnf4TransItemActionUPGRADE:
		return dnf5TransItemActionUpgrade, nil
	case dnf4TransItemActionREMOVE:
		return dnf5TransItemActionRemove, nil
	case dnf4TransItemActionREINSTALL:
		return dnf5TransItemActionReinstall, nil
	case dnf4TransItemActionREASON_CHANGE:
		return dnf5TransItemActionReasonChange, nil
	// commit 02f66e31fbc315d8b515b84c0a82770635a7c663
	// Replace reverse transaction actions by REPLACED
	case dnf4TransItemActionDOWNGRADED, dnf4TransItemActionOBSOLETED,
		dnf4TransItemActionUPGRADED, dnf4TransItemActionREINSTALLED:
		return dnf5TransItemActionReplaced, nil
	// commit 58dcaa94f4e1038cdcc5ee7861633694534745bf
	// Quote from jmracek: OBSOLETE action is not used in DNF-4 and DNF-5.
	case dnf4TransItemActionOBSOLETE:
		fallthrough
	default:
		return 0, EnumError{"dnf4TransItemAction", int32(a)}
	}
}

func castTransItemReason(r dnf4TransItemReason) (dnf5TransItemReason, error) {
	switch r {
	// commit 9f506f40dddb2fea062196d1d4618e7166050238
	// libdnf/transaction/transaction_item_reason: Rename UNKNOWN reason to NONE
	case dnf4TransItemReasonUNKNOWN:
		// commit cfaa63f94e11dbdc788ed5852273087d8f7c52d2
		// libdnf/system/state: Store dependencies, store from_repo for nevras
		return dnf5TransItemReasonExternalUser, nil
	case dnf4TransItemReasonDEPENDENCY:
		return dnf5TransItemReasonDependency, nil
	case dnf4TransItemReasonUSER:
		return dnf5TransItemReasonUser, nil
	case dnf4TransItemReasonCLEAN:
		return dnf5TransItemReasonClean, nil
	case dnf4TransItemReasonWEAK_DEPENDENCY:
		return dnf5TransItemReasonWeakDependency, nil
	case dnf4TransItemReasonGROUP:
		return dnf5TransItemReasonGroup, nil
	default:
		return 0, EnumError{"dnf4TransItemReason", int32(r)}
	}
}

func castTransItemState(s dnf4TransItemState) (dnf5TransItemState, error) {
	// commit 75de6377056cceba44b84f83d118c293fc2b1706
	// libdnf/transaction: Rename transaction item states
	switch s {
	case dnf4TransItemStateUNKNOWN:
		return dnf5TransItemStateStarted, nil
	case dnf4TransItemStateDONE:
		return dnf5TransItemStateOk, nil
	case dnf4TransItemStateERROR:
		return dnf5TransItemStateError, nil
	default:
		return 0, EnumError{"dnf4TransItemState", int32(s)}
	}
}

func migrateTransItem(ctx context.Context, db4 *sql.DB, db5 *sql.DB) error {
	rows, err := db4.QueryContext(ctx, `SELECT
		id, trans_id, item_id, repo_id,
		action, reason, state
	FROM trans_item ORDER BY id ASC`)
	if err != nil {
		return FuncError{"db4.QueryContext", err}
	}
	defer rows.Close()

	return transaction(ctx, db5, func(tx5 *sql.Tx) error {
		for rows.Next() {
			r4 := dnf4TransItem{}
			err := rows.Scan(&r4.id, &r4.trans_id, &r4.item_id, &r4.repo_id,
				&r4.action, &r4.reason, &r4.state)
			if err != nil {
				return FuncError{"rows.Scan", err}
			}

			action, err := castTransItemAction(r4.action)
			if err != nil {
				return FuncError{"castTransItemAction", err}
			}
			reason, err := castTransItemReason(r4.reason)
			if err != nil {
				return FuncError{"castTransItemReason", err}
			}
			state, err := castTransItemState(r4.state)
			if err != nil {
				return FuncError{"castTransItemState", err}
			}
			r5 := dnf5TransItem{
				id:        r4.id,
				trans_id:  r4.trans_id,
				item_id:   r4.item_id,
				repo_id:   r4.repo_id,
				action_id: action,
				reason_id: reason,
				state_id:  state,
			}

			_, err = tx5.ExecContext(ctx, `INSERT INTO trans_item (
				id, trans_id, item_id, repo_id,
				action_id, reason_id, state_id
			) VALUES (?, ?, ?, ?, ?, ?, ?)`,
				r5.id, r5.trans_id, r5.item_id, r5.repo_id,
				r5.action_id, r5.reason_id, r5.state_id)
			if err != nil {
				return FuncError{"tx5.ExecContext", err}
			}
		}
		return rows.Err()
	})
}
