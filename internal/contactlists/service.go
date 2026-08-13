package contactlists

import (
	"errors"

	"gorm.io/gorm"
)

var ErrListNotFound = errors.New("contact list not found")

type ListWithCount struct {
	ContactList
	MemberCount int64 `json:"member_count"`
}

func List(db *gorm.DB, ownerID string) ([]ListWithCount, error) {
	var lists []ContactList
	if err := db.Where("owner_id = ?", ownerID).Order("created_at DESC").Find(&lists).Error; err != nil {
		return nil, err
	}

	out := make([]ListWithCount, len(lists))
	for i, l := range lists {
		var count int64
		db.Model(&ContactListMember{}).Where("list_id = ?", l.ID).Count(&count)
		out[i] = ListWithCount{ContactList: l, MemberCount: count}
	}
	return out, nil
}

func GetByID(db *gorm.DB, ownerID, id string) (*ContactList, error) {
	var l ContactList
	err := db.Where("id = ? AND owner_id = ?", id, ownerID).First(&l).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrListNotFound
	}
	return &l, err
}

type CreateInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func Create(db *gorm.DB, ownerID string, in *CreateInput) (*ContactList, error) {
	l := &ContactList{
		OwnerID:     ownerID,
		Name:        in.Name,
		Description: in.Description,
	}
	if err := db.Create(l).Error; err != nil {
		return nil, err
	}
	return l, nil
}

type UpdateInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func Update(db *gorm.DB, l *ContactList, in *UpdateInput) (*ContactList, error) {
	if in.Name != nil {
		l.Name = *in.Name
	}
	if in.Description != nil {
		l.Description = *in.Description
	}
	return l, db.Save(l).Error
}

func Delete(db *gorm.DB, l *ContactList) error {
	db.Where("list_id = ?", l.ID).Delete(&ContactListMember{})
	return db.Delete(l).Error
}

func GetMembers(db *gorm.DB, listID string) ([]string, error) {
	var members []ContactListMember
	if err := db.Where("list_id = ?", listID).Find(&members).Error; err != nil {
		return nil, err
	}
	ids := make([]string, len(members))
	for i, m := range members {
		ids[i] = m.ContactID
	}
	return ids, nil
}

type AddMembersInput struct {
	ContactIDs []string `json:"contact_ids" binding:"required"`
}

func AddMembers(db *gorm.DB, listID string, contactIDs []string) (int, error) {
	added := 0
	for _, cid := range contactIDs {
		var count int64
		db.Model(&ContactListMember{}).Where("list_id = ? AND contact_id = ?", listID, cid).Count(&count)
		if count > 0 {
			continue
		}
		m := &ContactListMember{ListID: listID, ContactID: cid}
		if err := db.Create(m).Error; err != nil {
			return added, err
		}
		added++
	}
	return added, nil
}

func RemoveMembers(db *gorm.DB, listID string, contactIDs []string) error {
	return db.Where("list_id = ? AND contact_id IN ?", listID, contactIDs).Delete(&ContactListMember{}).Error
}
